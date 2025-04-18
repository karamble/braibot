package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	// Keep for PM type reference if needed indirectly
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// ImageService handles image generation
type ImageService struct {
	client         *fal.Client
	dbManager      *database.DBManager
	bot            *kit.Bot
	debug          bool
	billingEnabled bool // Added billing enabled flag
}

// NewImageService creates a new ImageService
func NewImageService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool, billingEnabled bool) *ImageService {
	return &ImageService{
		client:         client,
		dbManager:      dbManager,
		bot:            bot,
		debug:          debug,
		billingEnabled: billingEnabled, // Store the flag
	}
}

// GenerateImage generates an image based on the request, handling billing after successful result sending.
func (s *ImageService) GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	// 1. Validate request
	if err := s.validateRequest(req); err != nil {
		return &ImageResult{Success: false, Error: err}, err
	}

	// 2. Calculate expected cost and CHECK balance (don't deduct yet)
	numImagesToRequest := req.NumImages
	if numImagesToRequest <= 0 {
		numImagesToRequest = 1 // Default to 1 if not specified or invalid
	}
	totalExpectedCostUSD := req.PriceUSD * float64(numImagesToRequest)

	// Check balance, passing the billingEnabled flag
	hasSufficientBalance, requiredDCR, currentBalanceDCR, insufErrMsg, checkErr := utils.CheckBalance(ctx, s.dbManager, req.UserID[:], totalExpectedCostUSD, s.debug, s.billingEnabled)
	if checkErr != nil {
		// Critical error during balance check
		err := fmt.Errorf("failed during balance check: %v", checkErr)
		return &ImageResult{Success: false, Error: err}, err
	}
	if !hasSufficientBalance {
		// Insufficient funds (only happens if billing is enabled)
		return &ImageResult{Success: false, Error: fmt.Errorf(insufErrMsg)}, nil
	}

	// Inform user about processing (adjust message based on billing)
	var infoMsg string
	if s.billingEnabled {
		infoMsg = fmt.Sprintf("Request cost: $%.2f USD (%.8f DCR). Your balance: %.8f DCR. Processing...", totalExpectedCostUSD, requiredDCR, currentBalanceDCR)
	} else {
		infoMsg = "Processing your request (billing disabled)..."
	}
	s.bot.SendPM(ctx, req.UserNick, infoMsg)

	// 4. Create the appropriate FAL request object using the helper function
	falReq, err := createFalImageRequest(req, numImagesToRequest)
	if err != nil {
		// Handle error from request creation (e.g., unsupported model)
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error creating generation request: %v", err))
		return &ImageResult{Success: false, Error: err}, err // No billing occurred
	}

	// 5. Generate image using the created request
	imageResp, genErr := s.client.GenerateImage(ctx, falReq)
	if genErr != nil {
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Image generation failed: %v", genErr))
		return &ImageResult{Success: false, Error: genErr}, genErr // No billing occurred
	}

	// 6. Check if the image URL is empty - check if *any* images were returned
	if len(imageResp.Images) == 0 {
		genErr = fmt.Errorf("API did not return any images")
		s.bot.SendPM(ctx, req.UserNick, genErr.Error())
		return &ImageResult{Success: false, Error: genErr}, genErr // No billing occurred
	}

	// 7. Send the image(s) - loop through results
	numImagesGenerated := len(imageResp.Images)
	successfullySentCount := 0
	var lastSentImageURL string // Keep track of the last URL for the result
	for i, img := range imageResp.Images {
		if img.URL == "" {
			s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Skipping image %d/%d: received empty URL from API.", i+1, numImagesGenerated))
			continue
		}
		lastSentImageURL = img.URL // Update last URL
		contentType := img.ContentType
		s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Sending image %d of %d...", i+1, numImagesGenerated))

		var sendErr error
		if strings.Contains(contentType, "svg") || !strings.HasPrefix(contentType, "image/") {
			// For SVG or non-standard image formats, use SendFile
			sendErr = utils.SendFileToUser(ctx, s.bot, req.UserNick, img.URL, "image", contentType)
		} else {
			// For standard image formats, use PM embed
			sendErr = sendEmbeddedImage(ctx, s.bot, req, img, i, numImagesGenerated)
		}

		if sendErr != nil {
			s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to send image %d/%d: %v", i+1, numImagesGenerated, sendErr))
			// Optionally continue to try sending other images
		} else {
			successfullySentCount++
		}
	}

	// Send seed information if available
	if imageResp.Seed != 0 {
		seedMsg := fmt.Sprintf("🌱 Seed for the request: %d", imageResp.Seed)
		if err := s.bot.SendPM(ctx, req.UserNick, seedMsg); err != nil {
			fmt.Printf("WARN: Failed to send seed message to %s: %v\n", req.UserNick, err)
		}
	}

	// 8. Perform Billing *only if* enabled and at least one image was sent successfully
	var chargedDCR float64
	var finalBalanceDCR float64 = currentBalanceDCR // Start with the balance known before potential deduction
	var billingAttempted bool = false
	var billingSucceeded bool = false

	if s.billingEnabled && successfullySentCount > 0 {
		billingAttempted = true
		deductChargedDCR, deductNewBalance, deductErr := utils.DeductBalance(ctx, s.dbManager, req.UserID[:], totalExpectedCostUSD, s.debug, s.billingEnabled)
		if deductErr != nil {
			// Log the error and inform the user
			s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error processing payment after sending results: %v. Please contact support.", deductErr))
			// Use the balance known before the failed deduction attempt
			finalBalanceDCR = currentBalanceDCR
		} else {
			billingSucceeded = true
			chargedDCR = deductChargedDCR
			finalBalanceDCR = deductNewBalance
		}
	} else if !s.billingEnabled {
		// fmt.Printf("INFO: Billing is disabled. No charge applied for user %s.\n", req.UserNick) // Already Removed
	} else {
		// Billing enabled, but no images sent successfully
		// fmt.Printf("INFO: No images sent successfully for user %s. No billing occurred.\n", req.UserNick) // Removed
	}

	// 9. Send final confirmation
	finalMessage := fmt.Sprintf("Finished processing request. Sent %d of %d generated image(s).\n\n", successfullySentCount, numImagesGenerated)

	if s.billingEnabled {
		if billingAttempted && billingSucceeded {
			finalMessage += fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• New Balance: %.8f DCR",
				chargedDCR,
				totalExpectedCostUSD, // Using the original cost USD for consistency
				finalBalanceDCR)
		} else if billingAttempted && !billingSucceeded {
			// Images sent, but billing failed
			finalMessage += fmt.Sprintf("⚠️ Billing failed after sending results. Your balance remains %.8f DCR. Please contact support.", finalBalanceDCR)
		} else {
			// Billing enabled, but no charge was applied (e.g., no images sent)
			finalMessage += fmt.Sprintf("No charge was applied. Your balance remains %.8f DCR.", finalBalanceDCR)
		}
	} else {
		// Billing is disabled
		finalMessage += "Billing is disabled. No charge was applied."
	}

	if err := s.bot.SendPM(ctx, req.UserNick, finalMessage); err != nil {
		// Log error, but don't fail the whole operation just because the final message failed
		// fmt.Printf("ERROR: Failed to send final confirmation message to %s: %v\n", req.UserNick, err) // Removed
	}

	// Return success if at least one image was generated, using the last URL
	if numImagesGenerated > 0 {
		// Indicate overall success based on generation, even if sending/billing had issues
		// The final message informs the user about those issues.
		return &ImageResult{
			ImageURL: lastSentImageURL, // Return the URL of the last image generated/sent
			Success:  true,             // Represents successful generation from the API
		}, nil
	} else {
		// This case should ideally be caught earlier, but as a fallback
		return &ImageResult{Success: false, Error: fmt.Errorf("no images were generated successfully")}, nil
	}
}

// sendEmbeddedImage fetches, encodes, and sends an image embedded in a PM.
func sendEmbeddedImage(ctx context.Context, bot *kit.Bot, req *ImageRequest, img fal.ImageOutput, index, total int) error {
	// Fetch the image data
	imgDataResp, err := http.Get(img.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch image %d/%d: %w", index+1, total, err)
	}
	defer imgDataResp.Body.Close()

	if imgDataResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch image %d/%d: status %s", index+1, total, imgDataResp.Status)
	}

	imageData, err := io.ReadAll(imgDataResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image data %d/%d: %w", index+1, total, err)
	}

	// Encode the image data to base64
	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	// Create the message with embedded image
	message := fmt.Sprintf("--embed[alt=%s image %d/%d,type=%s,data=%s]--",
		req.ModelName,
		index+1,
		total,
		img.ContentType,
		encodedImage)

	return bot.SendPM(ctx, req.UserNick, message)
}

// Helper function to safely dereference optional int pointers
func derefIntPtrOrDefault(ptr *int, defaultValue int) int {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// validateRequest validates the image request
func (s *ImageService) validateRequest(req *ImageRequest) error {
	// Check if model exists
	_, exists := faladapter.GetCurrentModel(req.ModelType)
	if !exists {
		return fmt.Errorf("no default model found for %s", req.ModelType)
	}

	// For image2image, check if image URL is provided
	if req.ModelType == "image2image" && req.ImageURL == "" {
		return fmt.Errorf("image URL is required for image2image")
	}

	return nil
}

// createFalImageRequest constructs the appropriate fal.Model request struct based on the internal ImageRequest.
func createFalImageRequest(req *ImageRequest, numImagesToRequest int) (interface{}, error) {
	var falReq interface{}

	// Create the specific fal request based on the model name
	switch req.ModelName {
	case "fast-sdxl":
		falReq = &fal.FastSDXLRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt,
				Progress: req.Progress,
			},
			// fast-sdxl specific options parsed from req if added
			NumImages: numImagesToRequest, // Use requested number
		}
	case "ghiblify":
		if req.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for ghiblify model")
		}
		falReq = &fal.GhiblifyRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt, // Optional prompt for ghiblify
				ImageURL: req.ImageURL,
				Progress: req.Progress,
			},
		}
	case "cartoonify":
		if req.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for cartoonify model")
		}
		falReq = &fal.CartoonifyRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt, // Allow optional prompt?
				ImageURL: req.ImageURL,
				Progress: req.Progress,
			},
		}
	case "star-vector":
		if req.ImageURL == "" {
			return nil, fmt.Errorf("image_url is required for star-vector model")
		}
		falReq = &fal.StarVectorRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt,
				ImageURL: req.ImageURL,
				Progress: req.Progress,
			},
		}
	case "flux/schnell":
		falReq = &fal.FluxSchnellRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt,
				Progress: req.Progress,
			},
			NumImages:           numImagesToRequest,
			ImageSize:           req.ImageSize,
			Seed:                req.Seed,
			NumInferenceSteps:   derefIntPtrOrDefault(req.NumInferenceSteps, 4),
			EnableSafetyChecker: req.EnableSafetyChecker,
		}
	case "flux-pro/v1.1":
		falReq = &fal.FluxProV1_1Request{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt,
				Progress: req.Progress,
			},
			NumImages:           numImagesToRequest,
			ImageSize:           req.ImageSize,
			Seed:                req.Seed,
			EnableSafetyChecker: req.EnableSafetyChecker,
			SafetyTolerance:     req.SafetyTolerance,
			OutputFormat:        req.OutputFormat,
		}
	case "flux-pro/v1.1-ultra":
		falReq = &fal.FluxProV1_1UltraRequest{
			BaseImageRequest: fal.BaseImageRequest{
				Prompt:   req.Prompt,
				Progress: req.Progress,
			},
			NumImages:           numImagesToRequest,
			Seed:                req.Seed,
			EnableSafetyChecker: req.EnableSafetyChecker,
			SafetyTolerance:     req.SafetyTolerance,
			OutputFormat:        req.OutputFormat,
			AspectRatio:         req.AspectRatio,
			Raw:                 req.Raw,
		}
	case "hidream-i1-full":
		falReq = &fal.HiDreamI1FullRequest{
			BaseImageRequest:    fal.BaseImageRequest{Prompt: req.Prompt, Progress: req.Progress},
			NegativePrompt:      req.NegativePrompt,
			ImageSize:           req.ImageSize,
			NumInferenceSteps:   req.NumInferenceSteps,
			Seed:                req.Seed,
			GuidanceScale:       req.GuidanceScale,
			NumImages:           numImagesToRequest,
			EnableSafetyChecker: req.EnableSafetyChecker,
			OutputFormat:        req.OutputFormat,
		}
	case "hidream-i1-dev":
		falReq = &fal.HiDreamI1DevRequest{
			HiDreamI1FullRequest: fal.HiDreamI1FullRequest{
				BaseImageRequest:    fal.BaseImageRequest{Prompt: req.Prompt, Progress: req.Progress},
				NegativePrompt:      req.NegativePrompt,
				ImageSize:           req.ImageSize,
				NumInferenceSteps:   req.NumInferenceSteps,
				Seed:                req.Seed,
				GuidanceScale:       req.GuidanceScale,
				NumImages:           numImagesToRequest,
				EnableSafetyChecker: req.EnableSafetyChecker,
				OutputFormat:        req.OutputFormat,
			},
		}
	case "hidream-i1-fast":
		falReq = &fal.HiDreamI1FastRequest{
			HiDreamI1FullRequest: fal.HiDreamI1FullRequest{
				BaseImageRequest:    fal.BaseImageRequest{Prompt: req.Prompt, Progress: req.Progress},
				NegativePrompt:      req.NegativePrompt,
				ImageSize:           req.ImageSize,
				NumInferenceSteps:   req.NumInferenceSteps,
				Seed:                req.Seed,
				GuidanceScale:       req.GuidanceScale,
				NumImages:           numImagesToRequest,
				EnableSafetyChecker: req.EnableSafetyChecker,
				OutputFormat:        req.OutputFormat,
			},
		}
	// Add cases for other specific image models here
	default:
		return nil, fmt.Errorf("unsupported or unhandled model for specific FAL image request creation: %s", req.ModelName)
	}
	return falReq, nil
}
