package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// ImageService handles image generation
type ImageService struct {
	client    *fal.Client
	dbManager *database.DBManager
	bot       *kit.Bot
	debug     bool
}

// NewImageService creates a new ImageService
func NewImageService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool) *ImageService {
	return &ImageService{
		client:    client,
		dbManager: dbManager,
		bot:       bot,
		debug:     debug,
	}
}

// GenerateImage generates an image based on the request
func (s *ImageService) GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	// 1. Validate request
	if err := s.validateRequest(req); err != nil {
		return &ImageResult{Success: false, Error: err}, err
	}

	// 2. Calculate expected cost and check balance
	numImagesToRequest := req.NumImages
	if numImagesToRequest <= 0 {
		numImagesToRequest = 1 // Default to 1 if not specified or invalid
	}
	totalExpectedCostUSD := req.PriceUSD * float64(numImagesToRequest)

	pm := types.ReceivedPM{
		Nick: req.UserNick,
		Uid:  req.UserID[:],
	}
	billingResult, billingErr := utils.CheckAndProcessBilling(ctx, s.bot, s.dbManager, pm, totalExpectedCostUSD, s.debug)
	if billingErr != nil {
		return &ImageResult{Success: false, Error: billingErr}, billingErr
	}
	if !billingResult.Success {
		return &ImageResult{Success: false, Error: fmt.Errorf(billingResult.ErrorMessage)}, nil
	}

	// 3. Send initial message
	s.bot.SendPM(ctx, req.UserNick, "Processing your request.")

	// 4. Create the appropriate FAL request object using the helper function
	falReq, err := createFalImageRequest(req, numImagesToRequest)
	if err != nil {
		// Handle error from request creation (e.g., unsupported model)
		// Attempt to refund if billing was successful but we can't create the request
		// if refundErr := utils.RefundBilling(ctx, s.bot, pm, billingResult); refundErr != nil {
		// 	s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to process refund after request creation error: %v", refundErr))
		// }
		return &ImageResult{Success: false, Error: err}, err
	}

	// 5. Generate image using the created request
	imageResp, genErr := s.client.GenerateImage(ctx, falReq)
	if genErr != nil {
		// Attempt to refund if billing was successful but generation failed
		// Note: Refund logic might need adjustment based on CheckAndProcessBilling behavior
		// if err := utils.RefundBilling(ctx, s.bot, pm, billingResult); err != nil {
		// 	 s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to process refund after generation error: %v", err))
		// }
		return &ImageResult{Success: false, Error: genErr}, genErr
	}

	// 6. Check if the image URL is empty - check if *any* images were returned
	if len(imageResp.Images) == 0 {
		genErr = fmt.Errorf("API did not return any images")
		// Attempt to refund if billing was successful but generation failed
		// Note: Refund logic might need adjustment based on CheckAndProcessBilling behavior
		// if err := utils.RefundBilling(ctx, s.bot, pm, billingResult); err != nil {
		// 	 s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to process refund after generation error: %v", err))
		// }
		return &ImageResult{Success: false, Error: genErr}, genErr
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

		if strings.Contains(contentType, "svg") || !strings.HasPrefix(contentType, "image/") {
			// For SVG or non-standard image formats, use SendFile
			if err := utils.SendFileToUser(ctx, s.bot, req.UserNick, img.URL, "image", contentType); err != nil {
				s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to send image %d/%d: %v", i+1, numImagesGenerated, err))
				// Optionally continue to try sending other images
			} else {
				successfullySentCount++
			}
		} else {
			// For standard image formats, use PM embed
			// Fetch the image data
			imgDataResp, err := http.Get(img.URL)
			if err != nil {
				s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to fetch image %d/%d: %v", i+1, numImagesGenerated, err))
				continue // Skip this image
			}
			defer imgDataResp.Body.Close()

			imageData, err := io.ReadAll(imgDataResp.Body)
			if err != nil {
				s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to read image data %d/%d: %v", i+1, numImagesGenerated, err))
				continue // Skip this image
			}

			// Encode the image data to base64
			encodedImage := base64.StdEncoding.EncodeToString(imageData)

			// Create the message with embedded image
			message := fmt.Sprintf("--embed[alt=%s image %d/%d,type=%s,data=%s]--",
				req.ModelName,
				i+1,
				numImagesGenerated,
				contentType,
				encodedImage)
			if err := s.bot.SendPM(ctx, req.UserNick, message); err != nil {
				s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to send embedded image %d/%d: %v", i+1, numImagesGenerated, err))
				// Optionally continue
			} else {
				successfullySentCount++
			}
		}
	}

	// Send seed information if available
	if imageResp.Seed != 0 {
		seedMsg := fmt.Sprintf("🌱 Seed for the request: %d", imageResp.Seed)
		if err := s.bot.SendPM(ctx, req.UserNick, seedMsg); err != nil {
			fmt.Printf("WARN: Failed to send seed message to %s: %v\n", req.UserNick, err)
		}
	}

	// 8. Send final confirmation and billing info
	finalMessage := fmt.Sprintf("Finished processing request. Sent %d of %d generated image(s).\n\n", successfullySentCount, numImagesGenerated)

	// Append billing info (using data from the initial check)
	finalMessage += fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
		billingResult.ChargedDCR,
		billingResult.ChargedUSD,
		billingResult.NewBalance)

	if err := s.bot.SendPM(ctx, req.UserNick, finalMessage); err != nil {
		// Log error, but don't fail the whole operation just because the final message failed
		fmt.Printf("ERROR: Failed to send final billing message to %s: %v\n", req.UserNick, err)
	}

	// Return success if at least one image was generated, using the last URL
	if numImagesGenerated > 0 {
		return &ImageResult{
			ImageURL: lastSentImageURL, // Return the URL of the last image generated/sent
			Success:  true,
		}, nil
	} else {
		// This case should ideally be caught earlier, but as a fallback
		return &ImageResult{Success: false, Error: fmt.Errorf("no images were generated successfully")}, nil
	}
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
