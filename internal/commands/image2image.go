package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2ImageCommand returns the image2image command
func Image2ImageCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2image")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2image",
			Description: "Transform an image using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !image2image [image_url]", model.Description)

	return Command{
		Name:        "image2image",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("image2image")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2image [image_url]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2image [image_url]")
			}

			imageURL := args[0]

			// Validate URL
			parsedURL, err := url.Parse(imageURL)
			if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				return bot.SendPM(ctx, pm.Nick, "Please provide a valid http:// or https:// URL for the image.")
			}

			// Validate that URL points to an image by making a HEAD request
			httpResp, err := http.Head(imageURL)
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Failed to access the URL: %v", err))
			}
			defer httpResp.Body.Close()

			inputContentType := httpResp.Header.Get("Content-Type")
			if !strings.HasPrefix(inputContentType, "image/") {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("The URL does not point to an image (Content-Type: %s). Please provide a valid image URL.", inputContentType))
			}

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("image2image")
			if !exists {
				return fmt.Errorf("no default model found for image2image")
			}

			// Process billing
			billingResult, err := utils.CheckAndProcessBilling(ctx, bot, dbManager, pm, model.PriceUSD, debug)
			if err != nil {
				return fmt.Errorf("billing error: %v", err)
			}
			if !billingResult.Success {
				return bot.SendPM(ctx, pm.Nick, billingResult.ErrorMessage)
			}

			// Send confirmation message
			bot.SendPM(ctx, pm.Nick, "Processing your request.")

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "image2image")

			// Generate image using the faladapter
			imageResp, err := faladapter.GenerateImageFromImage(ctx, client, "", imageURL, model.Name, progress)
			if err != nil {
				return fmt.Errorf("failed to generate image: %v", err)
			}

			// Check if the image URL is empty
			if len(imageResp.Images) == 0 || imageResp.Images[0].URL == "" {
				return fmt.Errorf("received empty image URL from API")
			}

			// Check if the content type is SVG or another non-standard image format
			outputContentType := imageResp.Images[0].ContentType
			if strings.Contains(outputContentType, "svg") || !strings.HasPrefix(outputContentType, "image/") {
				// For SVG or non-standard image formats, use SendFile
				if err := utils.SendFileToUser(ctx, bot, pm.Nick, imageResp.Images[0].URL, "image", outputContentType); err != nil {
					return fmt.Errorf("failed to send image file: %v", err)
				}
			} else {
				// For standard image formats, use PM embed
				// Fetch the image data
				finalResp, err := http.Get(imageResp.Images[0].URL)
				if err != nil {
					return fmt.Errorf("failed to fetch image: %v", err)
				}
				defer finalResp.Body.Close()

				imageData, err := io.ReadAll(finalResp.Body)
				if err != nil {
					return fmt.Errorf("failed to read image data: %v", err)
				}

				// Encode the image data to base64
				encodedImage := base64.StdEncoding.EncodeToString(imageData)

				// Create the message with embedded image
				message := fmt.Sprintf("--embed[alt=%s image,type=%s,data=%s]--",
					model.Name,
					outputContentType,
					encodedImage)
				if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
					return fmt.Errorf("failed to send image: %v", err)
				}
			}

			// Send billing information
			return utils.SendBillingMessage(ctx, bot, pm, billingResult)
		},
	}
}
