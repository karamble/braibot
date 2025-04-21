package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/utils"
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
		Category:    "🎨 AI Generation",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model
				model, exists := faladapter.GetCurrentModel("image2video")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Error: Default image2video model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("image2video", model, userID, dbManager)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !image2video [image_url] [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return bot.SendPM(ctx, pm.Nick, header+helpDoc)
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			prompt, imageURL, duration, aspectRatio, negativePrompt, cfgScalePtr, promptOptimizerPtr, err := parser.Parse(args, true) // Expect Image URL, ignore promptOptimizer
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Argument error: %v", err))
			}
			if imageURL == "" { // Image URL is required for image2video
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL as the first argument.")
			}
			if prompt == "" {
				// Prompt might be optional for some image2video models, but let's require it for now
				// unless specific models indicate otherwise.
				return bot.SendPM(ctx, pm.Nick, "Please provide a text prompt describing the desired animation.")
			}

			// Get model configuration (required for PriceUSD and logic)
			model, exists := faladapter.GetCurrentModel("image2video")
			if !exists {
				return fmt.Errorf("no default model found for image2video")
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
				Prompt:          prompt,
				Duration:        duration,
				AspectRatio:     aspectRatio,
				NegativePrompt:  negativePrompt,
				CFGScale:        cfgScalePtr,        // Assign the parsed pointer
				PromptOptimizer: promptOptimizerPtr, // Assign the parsed pointer (may be nil)
				ModelType:       "image2video",
				Progress:        progress,
				UserNick:        pm.Nick,
				UserID:          userID,
				PriceUSD:        model.PriceUSD,
			}

			// Set the correct image URL field based on the model
			if model.Name == "minimax/video-01-subject-reference" {
				req.SubjectReferenceImageURL = imageURL // Parsed URL is the subject reference
			} else {
				req.ImageURL = imageURL // For other models, it's the standard image URL
			}

			// Generate video using the service
			result, err := videoService.GenerateVideo(ctx, req)

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, pm, "image2video", result, err); handleErr != nil {
				return handleErr // Propagate error if not handled by the utility function
			}

			// If we reach here, the operation was successful and errors were handled
			return nil
		},
	}
}
