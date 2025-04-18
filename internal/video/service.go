package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// VideoService handles video generation
type VideoService struct {
	client    *fal.Client
	dbManager *database.DBManager
	bot       *kit.Bot
	debug     bool
}

// NewVideoService creates a new VideoService
func NewVideoService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool) *VideoService {
	return &VideoService{
		client:    client,
		dbManager: dbManager,
		bot:       bot,
		debug:     debug,
	}
}

// GenerateVideo generates a video based on the request
func (s *VideoService) GenerateVideo(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	// 1. Validate request
	if err := s.validateRequest(req); err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	// 2. Check if user has sufficient balance
	pm := types.ReceivedPM{
		Nick: req.UserNick,
		Uid:  req.UserID[:], // Convert zkidentity.ShortID to []byte
	}
	billingResult, err := utils.CheckAndProcessBilling(ctx, s.bot, s.dbManager, pm, req.PriceUSD, s.debug)
	if err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}
	if !billingResult.Success {
		return &VideoResult{Success: false, Error: fmt.Errorf(billingResult.ErrorMessage)}, nil
	}

	// 3. Send initial message
	s.bot.SendPM(ctx, req.UserNick, "Processing your request.")

	// 4. Get current model name
	model, exists := faladapter.GetCurrentModel(req.ModelType)
	if !exists {
		return &VideoResult{Success: false, Error: fmt.Errorf("no default model found for %s", req.ModelType)}, nil
	}

	// 5. Create the appropriate FAL request object using the helper function
	falReq, err := createFalVideoRequest(req, model.Name)
	if err != nil {
		// Handle error from request creation (e.g., unsupported model)
		// Optional: Add refund logic here if needed
		return &VideoResult{Success: false, Error: err}, err
	}

	// 6. Generate video using the created request
	videoResp, err := s.client.GenerateVideo(ctx, falReq)
	if err != nil {
		// Optional: Add refund logic here if needed
		return &VideoResult{Success: false, Error: err}, err
	}

	// 7. Download and send video
	videoURL := videoResp.GetURL()
	if videoURL == "" {
		err = fmt.Errorf("no video URL found in response")
		return &VideoResult{Success: false, Error: err}, err
	}

	if err := s.downloadAndSendVideo(ctx, req.UserNick, videoURL); err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	// 8. Send billing info
	if err := utils.SendBillingMessage(ctx, s.bot, pm, billingResult); err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	return &VideoResult{
		VideoURL: videoURL,
		Success:  true,
	}, nil
}

// validateRequest validates the video request
func (s *VideoService) validateRequest(req *VideoRequest) error {
	// Check if model exists
	_, exists := faladapter.GetCurrentModel(req.ModelType)
	if !exists {
		return fmt.Errorf("no default model found for %s", req.ModelType)
	}

	// Option validation is now handled within the fal.GenerateVideo function

	// For image2video, check if image URL is provided
	if req.ModelType == "image2video" && req.ImageURL == "" {
		return fmt.Errorf("image URL is required for image2video")
	}

	return nil
}

// downloadAndSendVideo downloads a video from a URL, sends it to the user, and cleans up
func (s *VideoService) downloadAndSendVideo(ctx context.Context, userNick string, videoURL string) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temp file when done

	// Download the video
	resp, err := http.Get(videoURL)
	if err != nil {
		return fmt.Errorf("failed to download video: %v", err)
	}
	defer resp.Body.Close()

	// Copy the video data to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save video: %v", err)
	}

	// Close the file before sending
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Send the file to the user
	if err := s.bot.SendFile(ctx, userNick, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to send video file: %v", err)
	}

	return nil
}

// Helper function to safely dereference optional float64 pointers
func derefFloat64PtrOrDefault(ptr *float64, defaultValue float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// createFalVideoRequest constructs the appropriate fal.Model request struct based on the internal VideoRequest.
func createFalVideoRequest(req *VideoRequest, modelName string) (interface{}, error) {
	base := fal.BaseVideoRequest{
		Prompt:   req.Prompt,
		ImageURL: req.ImageURL, // May be empty for text2video
		Progress: req.Progress,
		Options:  make(map[string]interface{}), // Initialize options map
	}

	// Create the specific fal request based on the model name
	switch modelName {
	case "kling-video-text", "kling-video-image":
		// For Kling, CFGScale comes from the internal request if set
		cfgScale := derefFloat64PtrOrDefault(req.CFGScale, 0.5) // Default from KlingVideoOptions

		falReq := &fal.KlingVideoRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration,
			AspectRatio:      req.AspectRatio,
			NegativePrompt:   req.NegativePrompt,
			CFGScale:         cfgScale,
		}
		// Adjust base fields specific to type if necessary
		if modelName == "kling-video-text" {
			falReq.BaseVideoRequest.ImageURL = "" // Ensure empty for text2video
		}
		return falReq, nil
	case "veo2":
		if base.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for veo2 model")
		}
		falReq := &fal.Veo2Request{
			BaseVideoRequest: base,
			Duration:         req.Duration,
			AspectRatio:      req.AspectRatio,
		}
		return falReq, nil
	// Add cases for other specific video models here
	default:
		return nil, fmt.Errorf("unsupported or unhandled model for specific FAL video request creation: %s", modelName)
	}
}
