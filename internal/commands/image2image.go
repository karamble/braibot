package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2ImageCommand returns the image2image command
func Image2ImageCommand(debug bool) Command {
	return Command{
		Name:        "image2image",
		Description: "Transforms an image using AI. Usage: !image2image [image_url]",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2image [image_url]")
			}

			imageURL := args[0]

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("image2image")
			if !exists {
				return fmt.Errorf("no default model found for image2image")
			}
			model, exists := falapi.GetModel(modelName, "image2image")
			if !exists {
				return fmt.Errorf("model not found: %s", modelName)
			}

			// Generate image
			ghiblifyResp, err := client.GenerateImageFromImage(ctx, "", imageURL, model.Name, bot, pm.Nick)
			if err != nil {
				return fmt.Errorf("failed to generate image: %v", err)
			}

			// Log the response for debugging
			if debug {
				fmt.Printf("Image transformation response: %+v\n", ghiblifyResp)
			}

			// Check if the image URL is empty
			if ghiblifyResp.Image.URL == "" {
				return fmt.Errorf("received empty image URL from API")
			}

			// Fetch the image data
			imgResp, err := http.Get(ghiblifyResp.Image.URL)
			if err != nil {
				return fmt.Errorf("failed to fetch image: %v", err)
			}
			defer imgResp.Body.Close()

			imgData, err := io.ReadAll(imgResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read image data: %v", err)
			}

			// Encode the image data to base64
			encodedImage := base64.StdEncoding.EncodeToString(imgData)

			// Create the message with embedded image
			message := fmt.Sprintf("--embed[alt=%s style transformation,type=%s,data=%s]--",
				model.Name,
				ghiblifyResp.Image.ContentType,
				encodedImage)
			return bot.SendPM(ctx, pm.Nick, message)
		},
	}
}
