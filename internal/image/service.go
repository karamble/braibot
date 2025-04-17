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

	// 2. Check if user has sufficient balance
	pm := types.ReceivedPM{
		Nick: req.UserNick,
		Uid:  req.UserID[:],
	}
	billingResult, billingErr := utils.CheckAndProcessBilling(ctx, s.bot, s.dbManager, pm, req.PriceUSD, s.debug)
	if billingErr != nil {
		return &ImageResult{Success: false, Error: billingErr}, billingErr
	}
	if !billingResult.Success {
		return &ImageResult{Success: false, Error: fmt.Errorf(billingResult.ErrorMessage)}, nil
	}

	// 3. Send initial message
	s.bot.SendPM(ctx, req.UserNick, "Processing your request.")

	// 4. Generate image
	var imageResp *fal.ImageResponse
	var genErr error

	if req.ModelType == "text2image" {
		// Create text2image request
		imageReq := fal.ImageRequest{
			Prompt:   req.Prompt,
			Model:    req.ModelName,
			Options:  map[string]interface{}{"num_images": 1},
			Progress: req.Progress,
		}
		imageResp, genErr = s.client.GenerateImage(ctx, imageReq)
	} else {
		// Create image2image request
		imageResp, genErr = faladapter.GenerateImageFromImage(ctx, s.client, req.Prompt, req.ImageURL, req.ModelName, req.Progress)
	}

	if genErr != nil {
		return &ImageResult{Success: false, Error: genErr}, genErr
	}

	// 5. Check if the image URL is empty
	if len(imageResp.Images) == 0 || imageResp.Images[0].URL == "" {
		genErr = fmt.Errorf("received empty image URL from API")
		return &ImageResult{Success: false, Error: genErr}, genErr
	}

	// 6. Send the image
	contentType := imageResp.Images[0].ContentType
	if strings.Contains(contentType, "svg") || !strings.HasPrefix(contentType, "image/") {
		// For SVG or non-standard image formats, use SendFile
		if err := utils.SendFileToUser(ctx, s.bot, req.UserNick, imageResp.Images[0].URL, "image", contentType); err != nil {
			return &ImageResult{Success: false, Error: err}, err
		}
	} else {
		// For standard image formats, use PM embed
		// Fetch the image data
		imageResp, err := http.Get(imageResp.Images[0].URL)
		if err != nil {
			return &ImageResult{Success: false, Error: err}, err
		}
		defer imageResp.Body.Close()

		imageData, err := io.ReadAll(imageResp.Body)
		if err != nil {
			return &ImageResult{Success: false, Error: err}, err
		}

		// Encode the image data to base64
		encodedImage := base64.StdEncoding.EncodeToString(imageData)

		// Create the message with embedded image
		message := fmt.Sprintf("--embed[alt=%s image,type=%s,data=%s]--",
			req.ModelName,
			contentType,
			encodedImage)
		if err := s.bot.SendPM(ctx, req.UserNick, message); err != nil {
			return &ImageResult{Success: false, Error: err}, err
		}
	}

	// 7. Send billing info
	if err := utils.SendBillingMessage(ctx, s.bot, pm, billingResult); err != nil {
		return &ImageResult{Success: false, Error: err}, err
	}

	return &ImageResult{
		ImageURL: imageResp.Images[0].URL,
		Success:  true,
	}, nil
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
