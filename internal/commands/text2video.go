package commands

import (
	"context"
	"fmt"
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

// Text2VideoCommand returns the text2video command
func Text2VideoCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2video")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "text2video",
			Description: "Generate a video from text using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !text2video [prompt] [--duration 5] [--aspect 16:9] [--negative_prompt \"blur, distort, and low quality\"] [--cfg_scale 0.5]", model.Description)

	return Command{
		Name:        "text2video",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("text2video")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2video [prompt] [--duration 5] [--aspect 16:9] [--negative_prompt \"blur, distort, and low quality\"] [--cfg_scale 0.5]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2video [prompt] [--duration 5] [--aspect 16:9] [--negative_prompt \"blur, distort, and low quality\"] [--cfg_scale 0.5]")
			}

			// Parse arguments
			prompt := ""
			rawDuration := "5"    // Default duration
			aspectRatio := "16:9" // Default aspect ratio
			negativePrompt := ""  // Default negative prompt
			cfgScale := 0.5       // Default CFG scale

			// Parse remaining arguments
			for i := 0; i < len(args); i++ {
				arg := args[i]
				if strings.HasPrefix(arg, "--") {
					// Handle flags
					switch arg {
					case "--duration":
						if i+1 < len(args) {
							rawDuration = args[i+1]
							i++
						}
					case "--aspect":
						if i+1 < len(args) {
							aspectRatio = args[i+1]
							i++
						}
					case "--negative_prompt":
						if i+1 < len(args) {
							negativePrompt = args[i+1]
							i++
						}
					case "--cfg_scale":
						if i+1 < len(args) {
							if scale, err := strconv.ParseFloat(args[i+1], 64); err == nil {
								cfgScale = scale
							}
							i++
						}
					}
				} else {
					// Collect prompt text
					if prompt == "" {
						prompt = arg
					} else {
						prompt += " " + arg
					}
				}
			}

			// Validate aspect ratio
			validAspectRatios := map[string]bool{
				"16:9": true,
				"9:16": true,
				"1:1":  true,
			}
			if !validAspectRatios[aspectRatio] {
				return bot.SendPM(ctx, pm.Nick, "Invalid aspect ratio. Must be one of: 16:9, 9:16, 1:1")
			}

			// Validate duration
			if rawDuration != "5" {
				return bot.SendPM(ctx, pm.Nick, "Invalid duration. Must be 5 seconds")
			}

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("text2video")
			if !exists {
				return fmt.Errorf("no default model found for text2video")
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
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2video")

			// Create video request
			req := &fal.TextToVideoRequest{
				Prompt:         prompt,
				Duration:       rawDuration,
				AspectRatio:    aspectRatio,
				NegativePrompt: negativePrompt,
				CFGScale:       cfgScale,
				Progress:       progress,
			}

			// Generate video
			resp, err := client.GenerateVideo(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate video: %v", err)
			}

			// Get video URL
			videoURL := resp.Video.URL
			if videoURL == "" {
				videoURL = resp.URL
			}
			if videoURL == "" {
				videoURL = resp.VideoURL
			}
			if videoURL == "" {
				return fmt.Errorf("no video URL found in response")
			}

			// Download and send video
			if err := downloadAndSendVideo(ctx, bot, pm.Nick, videoURL); err != nil {
				return fmt.Errorf("failed to send video: %v", err)
			}

			return nil
		},
	}
}
