package commands

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
	"github.com/vctt94/bisonbotkit/config"
)

// Text2ImageCommand returns the text2image command
func Text2ImageCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2image")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "text2image",
			Description: "Generate an image from text using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !text2image [prompt]", model.Description)

	return Command{
		Name:        "text2image",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("text2image")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2image [prompt]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2image [prompt]")
			}

			prompt := strings.Join(args, " ")

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("text2image")
			if !exists {
				return fmt.Errorf("no default model found for text2image")
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
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2image")

			// Create image request
			req := fal.ImageRequest{
				Prompt:   prompt,
				Model:    model.Name,
				Options:  map[string]interface{}{"num_images": 1},
				Progress: progress,
			}

			// Generate image
			resp, err := client.GenerateImage(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate image: %v", err)
			}

			// Check if the image URL is empty
			if len(resp.Images) == 0 || resp.Images[0].URL == "" {
				return fmt.Errorf("received empty image URL from API")
			}

			// Check if the content type is SVG or another non-standard image format
			contentType := resp.Images[0].ContentType
			if strings.Contains(contentType, "svg") || !strings.HasPrefix(contentType, "image/") {
				// For SVG or non-standard image formats, use SendFile
				if err := utils.SendFileToUser(ctx, bot, pm.Nick, resp.Images[0].URL, "image", contentType); err != nil {
					return fmt.Errorf("failed to send image file: %v", err)
				}
			} else {
				// For standard image formats, use PM embed
				// Fetch the image data
				imageResp, err := http.Get(resp.Images[0].URL)
				if err != nil {
					return fmt.Errorf("failed to fetch image: %v", err)
				}
				defer imageResp.Body.Close()

				imageData, err := io.ReadAll(imageResp.Body)
				if err != nil {
					return fmt.Errorf("failed to read image data: %v", err)
				}

				// Encode the image data to base64
				encodedImage := base64.StdEncoding.EncodeToString(imageData)

				// Create the message with embedded image
				message := fmt.Sprintf("--embed[alt=%s image,type=%s,data=%s]--",
					model.Name,
					contentType,
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
