package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/video"
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
			parser := video.NewArgumentParser()
			duration := parser.ParseDuration(args)
			aspectRatio := parser.ParseAspectRatio(args)
			negativePrompt := parser.ParseNegativePrompt(args)
			cfgScale := parser.ParseCFGScale(args)

			// Collect prompt text
			for i := 0; i < len(args); i++ {
				arg := args[i]
				if !strings.HasPrefix(arg, "--") {
					if prompt == "" {
						prompt = arg
					} else {
						prompt += " " + arg
					}
				} else {
					// Skip the value for flags
					i++
				}
			}

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("text2video")
			if !exists {
				return fmt.Errorf("no default model found for text2video")
			}

			// Create video service
			videoService := video.NewVideoService(client, dbManager, bot, debug)

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2video")

			// Create video request
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			req := &video.VideoRequest{
				Prompt:         prompt,
				Duration:       duration,
				AspectRatio:    aspectRatio,
				NegativePrompt: negativePrompt,
				CFGScale:       cfgScale,
				ModelType:      "text2video",
				Progress:       progress,
				UserNick:       pm.Nick,
				UserID:         userID,
				PriceUSD:       model.PriceUSD,
			}

			// Generate video
			result, err := videoService.GenerateVideo(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate video: %v", err)
			}

			if !result.Success {
				return fmt.Errorf("video generation failed: %v", result.Error)
			}

			return nil
		},
	}
}
