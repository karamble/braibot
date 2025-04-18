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

// Image2VideoCommand returns the image2video command
// It now requires a VideoService instance.
func Image2VideoCommand(dbManager *database.DBManager, videoService *video.VideoService, debug bool) Command {
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
	description := fmt.Sprintf("%s. Usage: !image2video [image_url] [prompt] [--duration 5] [--aspect 16:9]", model.Description)

	return Command{
		Name:        "image2video",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("image2video")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2video [image_url] [prompt] [--duration 5] [--aspect 16:9]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2video [image_url] [prompt] [--duration 5] [--aspect 16:9]")
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			prompt, imageURL, duration, aspectRatio, negativePrompt, cfgScalePtr, err := parser.Parse(args, true) // Expect Image URL
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Argument error: %v", err))
			}
			if prompt == "" {
				// Prompt might be optional for some image2video models, but let's require it for now
				// unless specific models indicate otherwise.
				return bot.SendPM(ctx, pm.Nick, "Please provide a text prompt describing the desired animation.")
			}

			// Don't create client here, use the one in the service
			// client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Video service is now passed in
			// videoService := video.NewVideoService(client, dbManager, bot, debug)

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "image2video")

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			req := &video.VideoRequest{
				Prompt:         prompt,
				ImageURL:       imageURL,
				Duration:       duration,
				AspectRatio:    aspectRatio,
				NegativePrompt: negativePrompt,
				CFGScale:       cfgScalePtr, // Assign the parsed pointer
				ModelType:      "image2video",
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
