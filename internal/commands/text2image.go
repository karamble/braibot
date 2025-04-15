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
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2ImageCommand returns the text2image command
func Text2ImageCommand(debug bool) Command {
	return Command{
		Name:        "text2image",
		Description: "Generates an image from text prompt. Usage: !text2image [prompt]",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) == 0 {
				return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2image [prompt]")
			}

			prompt := strings.Join(args, " ")

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("text2image")
			if !exists {
				return fmt.Errorf("no default model found for text2image")
			}
			model, exists := falapi.GetModel(modelName, "text2image")
			if !exists {
				return fmt.Errorf("model not found: %s", modelName)
			}

			// Generate image
			imageResp, err := client.GenerateImage(ctx, prompt, model.Name, bot, pm.Nick)
			if err != nil {
				return err
			}

			// Assuming the first image is the one we want to send
			if len(imageResp.Images) > 0 {
				imageURL := imageResp.Images[0].URL
				// Fetch the image data
				imgResp, err := http.Get(imageURL)
				if err != nil {
					return err
				}
				defer imgResp.Body.Close()

				imgData, err := io.ReadAll(imgResp.Body)
				if err != nil {
					return err
				}

				// Encode the image data to base64
				encodedImage := base64.StdEncoding.EncodeToString(imgData)

				// Determine the image type from ContentType
				var imageType string
				switch imageResp.Images[0].ContentType {
				case "image/jpeg":
					imageType = "image/jpeg"
				case "image/png":
					imageType = "image/png"
				case "image/webp":
					imageType = "image/webp"
				default:
					imageType = "image/jpeg" // Fallback to jpeg if unknown
				}

				// Create the message with embedded image, using the user's prompt as the alt text
				message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--", url.QueryEscape(prompt), imageType, encodedImage)
				return bot.SendPM(ctx, pm.Nick, message)
			} else {
				return bot.SendPM(ctx, pm.Nick, "No images were generated.")
			}
		},
	}
}
