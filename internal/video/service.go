package video

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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
	var requiredDCR, currentBalanceDCR float64
	var checkErr error
	if s.billingEnabled {
		// Call CheckBalance, which now returns the error directly if insufficient or other issue
		requiredDCR, currentBalanceDCR, checkErr = utils.CheckBalance(ctx, s.dbManager, req.UserID[:], req.PriceUSD, s.debug, s.billingEnabled)
		if checkErr != nil {
			// Return the error (could be ErrInsufficientBalance or another critical error)
			// The calling layer (main.go) will handle ErrInsufficientBalance specifically.
			return &VideoResult{Success: false, Error: checkErr}, checkErr
		}
	}

	// 3. Send initial message (adjusted for billing status)
	var infoMsg string
	if s.billingEnabled {
		infoMsg = fmt.Sprintf("Request cost: $%.2f USD (%.8f DCR). Your balance: %.8f DCR. Processing...", req.PriceUSD, requiredDCR, currentBalanceDCR)
	} else {
		infoMsg = "Processing your request (billing disabled)..."
	}
	if req.IsPM {
		s.bot.SendPM(ctx, req.UserID.String(), infoMsg)
	} else {
		s.bot.SendGC(ctx, req.GC, "Processing your video request...")
	}

	// 4. Get current model name
	var model faladapter.AppModel
	var exists bool
	if req.ModelName != "" {
		model, exists = faladapter.GetModel(req.ModelName, req.ModelType)
		if !exists {
			return &VideoResult{Success: false, Error: fmt.Errorf("model not found: %s", req.ModelName)}, nil
		}
	} else {
		model, exists = faladapter.GetCurrentModel(req.ModelType, "")
		if !exists {
			return &VideoResult{Success: false, Error: fmt.Errorf("no default model found for %s", req.ModelType)}, nil // No billing occurred
		}
	}

	// 5. Create the appropriate FAL request object using the helper function
	falReq, err := createFalVideoRequest(req, model.Name)
	if err != nil {
		// Handle error from request creation (e.g., unsupported model)
		utils.SendToUser(ctx, s.bot, req.IsPM, req.UserID.String(), req.GC, fmt.Sprintf("Error creating generation request: %v", err))
		return &VideoResult{Success: false, Error: err}, err // No billing occurred
	}

	// 6. Generate video using the created request
	videoResp, genErr := s.client.GenerateVideo(ctx, falReq)
	if genErr != nil {
		// Log error server-side, do not PM the user here.
		// Error will be handled by the command handler (logged and nil returned).
		// s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Video generation failed: %v", genErr))
		return &VideoResult{Success: false, Error: genErr}, genErr // Return error to command handler
	}

	// 7. Check if URL is present and attempt to send
	videoURL := videoResp.GetURL()
	if videoURL == "" {
		genErr = fmt.Errorf("API did not return a video URL")
		// Log error server-side, do not PM the user here.
		// Error will be handled by the command handler.
		// s.bot.SendPM(ctx, req.UserNick, genErr.Error())
		return &VideoResult{Success: false, Error: genErr}, genErr // Return error to command handler
	}

	successfullySent := false
	if err := s.downloadAndSendVideo(ctx, req.UserNick, videoURL); err != nil {
		fmt.Printf("ERROR [VideoService] User %s: Failed to download/send video: %v\n", req.UserNick, err)
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
			if req.IsPM {
				s.bot.SendPM(ctx, req.UserID.String(), fmt.Sprintf("Error processing payment after sending video: %v. Please contact support.", deductErr))
			}
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
	if req.IsPM {
		finalMessage += utils.FormatBillingConfirmation("video", s.billingEnabled, billingAttempted, billingSucceeded, chargedDCR, req.PriceUSD, finalBalanceDCR)
		if err := s.bot.SendPM(ctx, req.UserID.String(), finalMessage); err != nil {
			// fmt.Printf("ERROR: Failed to send final confirmation message (video) to %s: %v\n", req.UserNick, err) // Removed
		}
	} else {
		if err := s.bot.SendGC(ctx, req.GC, "Video generation completed."); err != nil {
			// fmt.Printf("ERROR: Failed to send final confirmation message (video) to GC %s: %v\n", req.GC, err) // Removed
		}
	}

	// Return overall success based on generation, even if sending/billing failed
	return &VideoResult{
		VideoURL: videoURL,
		Success:  true, // Represents successful generation
	}, nil
}

// validateRequest validates the video request and formats duration based on model
func (s *VideoService) validateRequest(req *VideoRequest) error {
	// Check if model exists and get its details
	model, exists := faladapter.GetCurrentModel(req.ModelType, "")
	if !exists {
		return fmt.Errorf("no default model found for %s", req.ModelType)
	}

	// Format duration based on model
	switch model.Name {
	case "veo2":
		// Ensure duration HAS 's' suffix for veo2
		if _, err := strconv.Atoi(req.Duration); err == nil { // Check if it's a plain number
			if !strings.HasSuffix(req.Duration, "s") {
				req.Duration += "s" // Modify in place
			}
		} else {
			// If it's not a plain number, maybe it already has 's' or is invalid?
			// Add more robust validation here if needed.
			if !strings.HasSuffix(req.Duration, "s") {
				// Or return an error: return fmt.Errorf("invalid duration format for veo2: %s", req.Duration)
				req.Duration += "s" // Modify in place
			}
		}
	case "kling-video-text", "kling-video-image",
		"kling-video-v3-text", "kling-video-v3-pro-text",
		"kling-video-v3-image", "kling-video-v3-pro-image",
		"kling-video-o3-text", "kling-video-o3-pro-text",
		"seedance-2.0-image", "seedance-2.0-text", "seedance-2.0-reference":
		// Ensure duration does NOT have 's' suffix for Kling / Seedance
		if strings.HasSuffix(req.Duration, "s") {
			req.Duration = strings.TrimSuffix(req.Duration, "s") // Modify in place
		}
		// Optional: Add validation that it's a number if needed
	}

	// For video2video, check if the required video URL field is provided
	if req.ModelType == "video2video" {
		switch model.Name {
		case "kling-video-o3-edit", "kling-video-o3-pro-edit":
			if req.VideoURL == "" {
				return fmt.Errorf("video URL is required for model %s", model.Name)
			}
		}
	}

	// For multi2video (reference-to-video), enforce per-modality limits and the "audio requires image/video" constraint.
	if req.ModelType == "multi2video" {
		switch model.Name {
		case "seedance-2.0-reference":
			totalRefs := len(req.ImageURLs) + len(req.VideoURLs) + len(req.AudioURLs)
			if totalRefs == 0 {
				return fmt.Errorf("at least one reference input (--image, --video, or --audio) is required for %s", model.Name)
			}
			if totalRefs > 12 {
				return fmt.Errorf("total reference files must not exceed 12 (got %d)", totalRefs)
			}
			if len(req.ImageURLs) > 9 {
				return fmt.Errorf("max 9 reference images (got %d)", len(req.ImageURLs))
			}
			if len(req.VideoURLs) > 3 {
				return fmt.Errorf("max 3 reference videos (got %d)", len(req.VideoURLs))
			}
			if len(req.AudioURLs) > 3 {
				return fmt.Errorf("max 3 reference audio files (got %d)", len(req.AudioURLs))
			}
			if len(req.AudioURLs) > 0 && len(req.ImageURLs)+len(req.VideoURLs) == 0 {
				return fmt.Errorf("reference audio requires at least one reference image or video")
			}
		}
	}

	// Option validation for other parameters is now handled within the fal.GenerateVideo function

	// For image2video, check if the required image URL field is provided based on the model
	if req.ModelType == "image2video" {
		switch model.Name {
		case "minimax/video-01-subject-reference":
			if req.SubjectReferenceImageURL == "" {
				return fmt.Errorf("subject_reference_image_url is required for model %s", model.Name)
			}
		// Add cases for other image2video models that might use different URL fields
		default:
			// Default check for models using the standard ImageURL field (e.g., veo2, kling-video-image)
			if req.ImageURL == "" {
				return fmt.Errorf("image URL is required for model %s", model.Name)
			}
		}
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

// Helper function to safely dereference optional bool pointers
func derefBoolPtrOrDefault(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// createFalVideoRequest constructs the appropriate fal.Model request struct based on the internal VideoRequest.
// Assumes req.Duration has already been formatted by validateRequest.
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

		// Duration formatting removed - handled in validateRequest
		falReq := &fal.KlingVideoRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration, // Use pre-formatted duration
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

		// Duration formatting removed - handled in validateRequest
		falReq := &fal.Veo2Request{
			BaseVideoRequest: base,
			Duration:         req.Duration, // Use pre-formatted duration
			AspectRatio:      req.AspectRatio,
		}
		return falReq, nil
	case "minimax/video-01-director":
		if base.ImageURL != "" {
			return nil, fmt.Errorf("image_url is not supported for %s model", modelName)
		}
		// Get the default PromptOptimizer value from the model definition
		model, _ := faladapter.GetModel(modelName, "text2video") // Ignore error as model should exist
		defaultOptimizer := true                                 // Default fallback
		if modelOpts, ok := model.Options.(*fal.MinimaxDirectorOptions); ok && modelOpts.PromptOptimizer != nil {
			defaultOptimizer = *modelOpts.PromptOptimizer
		}
		promptOptimizer := derefBoolPtrOrDefault(req.PromptOptimizer, defaultOptimizer)

		falReq := &fal.MinimaxDirectorRequest{
			BaseVideoRequest: base,
			PromptOptimizer:  &promptOptimizer,
		}
		falReq.BaseVideoRequest.ImageURL = "" // Ensure ImageURL is empty
		return falReq, nil
	case "minimax/video-01-subject-reference":
		if req.SubjectReferenceImageURL == "" {
			return nil, fmt.Errorf("subject_reference_image_url is required for %s model", modelName)
		}
		// Get the default PromptOptimizer value from the model definition
		model, _ := faladapter.GetModel(modelName, "image2video") // Ignore error as model should exist
		defaultOptimizer := true                                  // Default fallback
		if modelOpts, ok := model.Options.(*fal.MinimaxSubjectReferenceOptions); ok && modelOpts.PromptOptimizer != nil {
			defaultOptimizer = *modelOpts.PromptOptimizer
		}
		promptOptimizer := derefBoolPtrOrDefault(req.PromptOptimizer, defaultOptimizer)

		falReq := &fal.MinimaxSubjectReferenceRequest{
			BaseVideoRequest:         base, // Includes Prompt, Progress
			SubjectReferenceImageURL: req.SubjectReferenceImageURL,
			PromptOptimizer:          &promptOptimizer,
		}
		falReq.BaseVideoRequest.ImageURL = "" // Ensure base ImageURL is empty as it's not used
		return falReq, nil
	case "minimax/video-01-live":
		if req.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s model", modelName)
		}
		// Get the default PromptOptimizer value from the model definition
		model, _ := faladapter.GetModel(modelName, "image2video")
		defaultOptimizer := true
		if modelOpts, ok := model.Options.(*fal.MinimaxLiveOptions); ok && modelOpts.PromptOptimizer != nil {
			defaultOptimizer = *modelOpts.PromptOptimizer
		}
		promptOptimizer := derefBoolPtrOrDefault(req.PromptOptimizer, defaultOptimizer)

		falReq := &fal.MinimaxLiveRequest{
			BaseVideoRequest: base,
			PromptOptimizer:  &promptOptimizer,
		}
		return falReq, nil
	case "minimax/video-01":
		if base.ImageURL != "" {
			return nil, fmt.Errorf("image_url is not supported for %s model", modelName)
		}
		// Get the default PromptOptimizer value from the model definition
		model, _ := faladapter.GetModel(modelName, "text2video") // Ignore error as model should exist
		defaultOptimizer := true                                 // Default fallback
		if modelOpts, ok := model.Options.(*fal.MinimaxVideo01Options); ok && modelOpts.PromptOptimizer != nil {
			defaultOptimizer = *modelOpts.PromptOptimizer
		}
		promptOptimizer := derefBoolPtrOrDefault(req.PromptOptimizer, defaultOptimizer)

		falReq := &fal.MinimaxVideo01Request{
			BaseVideoRequest: base, // Includes Prompt, Progress
			PromptOptimizer:  &promptOptimizer,
		}
		falReq.BaseVideoRequest.ImageURL = "" // Ensure base ImageURL is empty
		return falReq, nil
	case "minimax/hailuo-02":
		if base.ImageURL != "" {
			return nil, fmt.Errorf("image_url is not supported for %s model", modelName)
		}
		// Get the default PromptOptimizer value from the model definition
		model, _ := faladapter.GetModel(modelName, "text2video") // Ignore error as model should exist
		defaultOptimizer := true                                 // Default fallback
		if modelOpts, ok := model.Options.(*fal.MinimaxHailuo02Options); ok && modelOpts.PromptOptimizer != nil {
			defaultOptimizer = *modelOpts.PromptOptimizer
		}
		promptOptimizer := derefBoolPtrOrDefault(req.PromptOptimizer, defaultOptimizer)

		falReq := &fal.MinimaxHailuo02Request{
			BaseVideoRequest: base, // Includes Prompt, Progress
			Duration:         req.Duration,
			PromptOptimizer:  &promptOptimizer,
		}
		falReq.BaseVideoRequest.ImageURL = "" // Ensure base ImageURL is empty
		return falReq, nil
	case "grok-imagine-video-text":
		if base.ImageURL != "" {
			return nil, fmt.Errorf("image_url is not supported for %s model", modelName)
		}
		duration := 0
		if req.Duration != "" {
			d, err := strconv.Atoi(req.Duration)
			if err == nil && d > 0 {
				duration = d
			}
		}

		falReq := &fal.GrokImagineVideoTextRequest{
			BaseVideoRequest: base,
			Duration:         duration,
			AspectRatio:      req.AspectRatio,
			Resolution:       req.Resolution,
		}
		falReq.BaseVideoRequest.ImageURL = ""
		return falReq, nil
	case "kling-video-v3-text", "kling-video-v3-pro-text":
		cfgScale := derefFloat64PtrOrDefault(req.CFGScale, 0.5)
		falReq := &fal.KlingVideoV3Request{
			BaseVideoRequest: base,
			Duration:         req.Duration,
			AspectRatio:      req.AspectRatio,
			NegativePrompt:   req.NegativePrompt,
			CFGScale:         cfgScale,
			GenerateAudio:    req.GenerateAudio,
		}
		falReq.BaseVideoRequest.Model = modelName
		falReq.BaseVideoRequest.ImageURL = "" // Ensure empty for text2video
		return falReq, nil
	case "kling-video-v3-image", "kling-video-v3-pro-image":
		if base.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s model", modelName)
		}
		cfgScale := derefFloat64PtrOrDefault(req.CFGScale, 0.5)
		falReq := &fal.KlingVideoV3Request{
			BaseVideoRequest: base,
			Duration:         req.Duration,
			AspectRatio:      req.AspectRatio,
			NegativePrompt:   req.NegativePrompt,
			CFGScale:         cfgScale,
			GenerateAudio:    req.GenerateAudio,
			EndImageURL:      req.EndImageURL,
		}
		falReq.BaseVideoRequest.Model = modelName
		return falReq, nil
	case "kling-video-o3-text", "kling-video-o3-pro-text":
		falReq := &fal.KlingVideoO3TextRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration,
			AspectRatio:      req.AspectRatio,
			GenerateAudio:    req.GenerateAudio,
		}
		falReq.BaseVideoRequest.Model = modelName
		falReq.BaseVideoRequest.ImageURL = "" // Ensure empty for text2video
		return falReq, nil
	case "seedance-2.0-image":
		if base.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for %s model", modelName)
		}
		falReq := &fal.SeedanceRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration, // Already stripped of 's' suffix by validateRequest
			AspectRatio:      req.AspectRatio,
			Resolution:       req.Resolution,
			GenerateAudio:    req.GenerateAudio,
			EndImageURL:      req.EndImageURL,
			Seed:             req.Seed,
			EndUserID:        req.UserID.String(), // ByteDance requires this for copyright tracking
		}
		falReq.BaseVideoRequest.Model = modelName
		return falReq, nil
	case "seedance-2.0-text":
		falReq := &fal.SeedanceRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration, // Already stripped of 's' suffix by validateRequest
			AspectRatio:      req.AspectRatio,
			Resolution:       req.Resolution,
			GenerateAudio:    req.GenerateAudio,
			Seed:             req.Seed,
			EndUserID:        req.UserID.String(), // ByteDance requires this for copyright tracking
		}
		falReq.BaseVideoRequest.Model = modelName
		falReq.BaseVideoRequest.ImageURL = "" // Ensure empty for text2video
		return falReq, nil
	case "seedance-2.0-reference":
		falReq := &fal.SeedanceReferenceRequest{
			BaseVideoRequest: base,
			Duration:         req.Duration, // Already stripped of 's' suffix by validateRequest
			AspectRatio:      req.AspectRatio,
			Resolution:       req.Resolution,
			GenerateAudio:    req.GenerateAudio,
			ImageURLs:        req.ImageURLs,
			VideoURLs:        req.VideoURLs,
			AudioURLs:        req.AudioURLs,
			Seed:             req.Seed,
			EndUserID:        req.UserID.String(), // ByteDance requires this for copyright tracking
		}
		falReq.BaseVideoRequest.Model = modelName
		falReq.BaseVideoRequest.ImageURL = "" // Not used; references come from the URL slices
		return falReq, nil
	case "kling-video-o3-edit", "kling-video-o3-pro-edit":
		if req.VideoURL == "" {
			return nil, fmt.Errorf("video_url is required for %s model", modelName)
		}
		falReq := &fal.KlingVideoO3EditRequest{
			BaseVideoRequest: base,
			VideoURL:         req.VideoURL,
			ImageURLs:        req.ImageURLs,
			KeepAudio:        req.KeepAudio,
		}
		falReq.BaseVideoRequest.Model = modelName
		falReq.BaseVideoRequest.ImageURL = "" // Not used for video2video edit
		return falReq, nil
	default:
		return nil, fmt.Errorf("unsupported or unhandled model for specific FAL video request creation: %s", modelName)
	}
}
