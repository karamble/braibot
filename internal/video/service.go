package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	// "github.com/companyzero/bisonrelay/clientrpc/types" // Only needed for the old billing call
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// VideoService handles video generation
type VideoService struct {
	client         *fal.Client
	dbManager      *database.DBManager
	bot            *kit.Bot
	debug          bool
	billingEnabled bool // Added billing enabled flag
}

// NewVideoService creates a new VideoService
func NewVideoService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool, billingEnabled bool) *VideoService {
	return &VideoService{
		client:         client,
		dbManager:      dbManager,
		bot:            bot,
		debug:          debug,
		billingEnabled: billingEnabled, // Store the flag
	}
}

// GenerateVideo generates a video based on the request, handling billing conditionally.
func (s *VideoService) GenerateVideo(ctx context.Context, req *VideoRequest) (*VideoResult, error) {
	// 1. Validate request
	if err := s.validateRequest(req); err != nil {
		return &VideoResult{Success: false, Error: err}, err
	}

	// 2. Calculate cost and CHECK balance if billing is enabled
	hasSufficientBalance := true
	var requiredDCR, currentBalanceDCR float64
	var checkErr error
	var insufErrMsg string
	if s.billingEnabled {
		hasSufficientBalance, requiredDCR, currentBalanceDCR, insufErrMsg, checkErr = utils.CheckBalance(ctx, s.dbManager, req.UserID[:], req.PriceUSD, s.debug, s.billingEnabled)
		if checkErr != nil {
			// Critical error during balance check
			err := fmt.Errorf("failed during balance check: %v", checkErr)
			return &VideoResult{Success: false, Error: err}, err
		}
		if !hasSufficientBalance {
			// Insufficient funds
			return &VideoResult{Success: false, Error: fmt.Errorf(insufErrMsg)}, nil
		}
	}

	// 3. Send initial message (adjusted for billing status)
	var infoMsg string
	if s.billingEnabled {
		infoMsg = fmt.Sprintf("Request cost: $%.2f USD (%.8f DCR). Your balance: %.8f DCR. Processing...", req.PriceUSD, requiredDCR, currentBalanceDCR)
	} else {
		infoMsg = "Processing your request (billing disabled)..."
	}
	s.bot.SendPM(ctx, req.UserNick, infoMsg)

	// 4. Get current model name
	model, exists := faladapter.GetCurrentModel(req.ModelType)
	if !exists {
		return &VideoResult{Success: false, Error: fmt.Errorf("no default model found for %s", req.ModelType)}, nil // No billing occurred
	}

	// 5. Create the appropriate FAL request object using the helper function
	falReq, err := createFalVideoRequest(req, model.Name)
	if err != nil {
		// Handle error from request creation (e.g., unsupported model)
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error creating generation request: %v", err))
		return &VideoResult{Success: false, Error: err}, err // No billing occurred
	}

	// 6. Generate video using the created request
	videoResp, genErr := s.client.GenerateVideo(ctx, falReq)
	if genErr != nil {
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Video generation failed: %v", genErr))
		return &VideoResult{Success: false, Error: genErr}, genErr // No billing occurred
	}

	// 7. Check if URL is present and attempt to send
	videoURL := videoResp.GetURL()
	if videoURL == "" {
		genErr = fmt.Errorf("API did not return a video URL")
		s.bot.SendPM(ctx, req.UserNick, genErr.Error())
		return &VideoResult{Success: false, Error: genErr}, genErr // No billing occurred
	}

	successfullySent := false
	if err := s.downloadAndSendVideo(ctx, req.UserNick, videoURL); err != nil {
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to send video: %v", err))
		// Continue to billing potentially? Or return error? For now, continue but don't bill.
	} else {
		successfullySent = true
	}

	// 8. Perform Billing *only if* enabled and video was sent successfully
	var chargedDCR float64
	var finalBalanceDCR float64 = currentBalanceDCR // Use balance from initial check
	var billingAttempted bool = false
	var billingSucceeded bool = false

	if s.billingEnabled && successfullySent {
		billingAttempted = true
		deductChargedDCR, deductNewBalance, deductErr := utils.DeductBalance(ctx, s.dbManager, req.UserID[:], req.PriceUSD, s.debug, s.billingEnabled)
		if deductErr != nil {
			s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error processing payment after sending video: %v. Please contact support.", deductErr))
			// Use pre-deduction balance
			finalBalanceDCR = currentBalanceDCR
		} else {
			billingSucceeded = true
			chargedDCR = deductChargedDCR
			finalBalanceDCR = deductNewBalance
		}
	} else if !s.billingEnabled {
		// fmt.Printf("INFO: Billing disabled. No charge for video for user %s.\n", req.UserNick) // Already Removed
	} else {
		// Billing enabled, but not sent successfully
		// fmt.Printf("INFO: Video not sent successfully for user %s. No billing occurred.\n", req.UserNick) // Removed
	}

	// 9. Send final confirmation
	finalMessage := "Finished processing video request.\n\n"
	if !successfullySent {
		finalMessage = "Video generation completed, but failed to send the result.\n\n"
	}

	if s.billingEnabled {
		if billingAttempted && billingSucceeded {
			finalMessage += fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• New Balance: %.8f DCR",
				chargedDCR, req.PriceUSD, finalBalanceDCR)
		} else if billingAttempted && !billingSucceeded {
			finalMessage += fmt.Sprintf("⚠️ Billing failed after sending video. Your balance remains %.8f DCR. Please contact support.", finalBalanceDCR)
		} else {
			finalMessage += fmt.Sprintf("No charge was applied. Your balance remains %.8f DCR.", finalBalanceDCR)
		}
	} else {
		finalMessage += "Billing is disabled. No charge was applied."
	}

	if err := s.bot.SendPM(ctx, req.UserNick, finalMessage); err != nil {
		// fmt.Printf("ERROR: Failed to send final confirmation message (video) to %s: %v\n", req.UserNick, err) // Removed
	}

	// Return overall success based on generation, even if sending/billing failed
	return &VideoResult{
		VideoURL: videoURL,
		Success:  true, // Represents successful generation
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
