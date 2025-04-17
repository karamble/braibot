package commands

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// downloadAndSendVideo downloads a video from a URL, sends it to the user, and cleans up
func downloadAndSendVideo(ctx context.Context, bot *kit.Bot, userNick string, videoURL string) error {
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
	if err := bot.SendFile(ctx, userNick, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to send video file: %v", err)
	}

	return nil
}

// Image2VideoCommand returns the image2video command
func Image2VideoCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2video")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2video",
			Description: "Generate a video from an image using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !image2video [image_url] [duration] [aspect_ratio] [negative_prompt] [cfg_scale]", model.Description)

	return Command{
		Name:        "image2video",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("image2video")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2video [image_url] [duration] [aspect_ratio] [negative_prompt] [cfg_scale]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2video [image_url] [duration] [aspect_ratio] [negative_prompt] [cfg_scale]")
			}

			// Parse arguments
			imageURL := args[0]
			duration := "5s"      // Default duration
			aspectRatio := "16:9" // Default aspect ratio
			negativePrompt := ""  // Default negative prompt
			cfgScale := 7.5       // Default CFG scale

			if len(args) > 1 {
				duration = args[1]
			}
			if len(args) > 2 {
				aspectRatio = args[2]
			}
			if len(args) > 3 {
				negativePrompt = args[3]
			}
			if len(args) > 4 {
				if scale, err := strconv.ParseFloat(args[4], 64); err == nil {
					cfgScale = scale
				}
			}

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("image2video")
			if !exists {
				return fmt.Errorf("no default model found for image2video")
			}

			// Calculate total price based on model and duration
			var totalPrice float64
			if model.Name == "kling-video" {
				// Parse duration for Kling
				dur, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
				if err != nil {
					dur = 5 // Default to 5 seconds if parsing fails
				}
				basePrice := model.PriceUSD // Base price for 5 seconds
				additionalSeconds := 0
				if dur > 5 {
					additionalSeconds = dur - 5
				}
				totalPrice = basePrice + (float64(additionalSeconds) * 0.4)
			} else if model.Name == "veo2" {
				// Parse duration for Veo2
				dur, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
				if err != nil {
					dur = 5 // Default to 5 seconds if parsing fails
				}
				basePrice := model.PriceUSD // Base price for 5 seconds
				additionalSeconds := 0
				if dur > 5 {
					additionalSeconds = dur - 5
				}
				totalPrice = basePrice + (float64(additionalSeconds) * 0.50)
			} else {
				// Fallback to model's base price
				totalPrice = model.PriceUSD
			}

			// Override the model price with the calculated price
			model.PriceUSD = totalPrice

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
			progress := NewCommandProgressCallback(bot, pm.Nick, "image2video")

			// Create video request based on model type
			var req interface{}
			switch model.Name {
			case "veo2":
				req = &fal.Veo2Request{
					BaseVideoRequest: fal.BaseVideoRequest{
						Prompt:   "",
						ImageURL: imageURL,
						Model:    model.Name,
						Progress: progress,
						Options:  make(map[string]interface{}),
					},
					Duration:    duration,
					AspectRatio: aspectRatio,
				}
			case "kling-video":
				req = &fal.KlingVideoRequest{
					BaseVideoRequest: fal.BaseVideoRequest{
						Prompt:   "",
						ImageURL: imageURL,
						Model:    model.Name,
						Progress: progress,
						Options:  make(map[string]interface{}),
					},
					Duration:       duration,
					AspectRatio:    aspectRatio,
					NegativePrompt: negativePrompt,
					CFGScale:       cfgScale,
				}
			default:
				return fmt.Errorf("unsupported model: %s", model.Name)
			}

			// Generate video
			resp, err := client.GenerateVideo(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate video: %v", err)
			}

			// Get the video URL from any of the possible fields
			videoURL := resp.GetURL()
			if videoURL == "" {
				return fmt.Errorf("received empty video URL from API")
			}

			// Send the video file to the user using the utility function
			if err := utils.SendFileToUser(ctx, bot, pm.Nick, videoURL, "video", "video/mp4"); err != nil {
				return fmt.Errorf("failed to send video file: %v", err)
			}

			// Send billing information
			return utils.SendBillingMessage(ctx, bot, pm, billingResult)
		},
	}
}
