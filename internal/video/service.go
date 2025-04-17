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

	// 4. Generate video
	var videoResp *fal.VideoResponse

	// Always use KlingVideoRequest, setting appropriate fields based on type
	// The underlying fal.GenerateVideo distinguishes model by name.
	videoReq := &fal.KlingVideoRequest{
		BaseVideoRequest: fal.BaseVideoRequest{
			Progress: req.Progress,                 // Pass progress callback
			Options:  make(map[string]interface{}), // Initialize options map
		},
		Duration:       req.Duration,
		AspectRatio:    req.AspectRatio,
		NegativePrompt: req.NegativePrompt,
		CFGScale:       req.CFGScale,
	}

	// Set fields based on ModelType
	if req.ModelType == "text2video" {
		videoReq.BaseVideoRequest.Model = "kling-video-text" // Specify the text model
		videoReq.BaseVideoRequest.Prompt = req.Prompt
	} else { // Assume image2video
		videoReq.BaseVideoRequest.Model = "kling-video-image" // Specify the image model (or could be veo2 later)
		videoReq.BaseVideoRequest.Prompt = req.Prompt         // Prompt might be used
		videoReq.BaseVideoRequest.ImageURL = req.ImageURL
	}

	// Generate video using the unified request type
	videoResp, err = s.client.GenerateVideo(ctx, videoReq)
	if err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	// 5. Download and send video
	videoURL := videoResp.GetURL()
	if videoURL == "" {
		err = fmt.Errorf("no video URL found in response")
		return &VideoResult{Success: false, Error: err}, err
	}

	if err := s.downloadAndSendVideo(ctx, req.UserNick, videoURL); err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	// 6. Send billing info
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

	// Validate options
	opts := &VideoOptions{
		Duration:       req.Duration,
		AspectRatio:    req.AspectRatio,
		NegativePrompt: req.NegativePrompt,
		CFGScale:       req.CFGScale,
	}

	parser := NewArgumentParser()
	if err := parser.ValidateOptions(opts); err != nil {
		return err
	}

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
