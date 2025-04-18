package commands

import (
	"context"
	"fmt"

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
// It now requires a VideoService instance.
func Text2VideoCommand(dbManager *database.DBManager, videoService *video.VideoService, debug bool) Command {
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

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			// Update variable list to match parser.Parse return values
			prompt, _, duration, aspectRatio, negativePrompt, cfgScalePtr, err := parser.Parse(args, false) // Expect NO Image URL
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Argument error: %v", err))
			}
			if prompt == "" {
				return bot.SendPM(ctx, pm.Nick, "Please provide a prompt text.")
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2video")

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			req := &video.VideoRequest{
				Prompt:         prompt,
				Duration:       duration,
				AspectRatio:    aspectRatio,
				NegativePrompt: negativePrompt,
				CFGScale:       cfgScalePtr, // Assign the parsed pointer
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
