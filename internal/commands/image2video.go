package commands

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/internal/video"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Image2VideoCommand returns the image2video command
// It now requires a VideoService instance.
func Image2VideoCommand(bot *kit.Bot, cfg *botconfig.BotConfig, imageService *video.VideoService, debug bool) braibottypes.Command {
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

	return braibottypes.Command{
		Name:        "image2video",
		Description: description,
		Category:    "🎨 AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Create a message sender using the adapter
			msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

			if len(args) < 1 {
				// Get the current model
				model, exists := faladapter.GetCurrentModel("image2video")
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default image2video model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("image2video", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !image2video [image_url] [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			prompt, imageURL, duration, aspectRatio, negativePrompt, cfgScalePtr, promptOptimizerPtr, err := parser.Parse(args, true) // Expect Image URL, ignore promptOptimizer
			if err != nil {
				return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Argument error: %v", err))
			}
			if imageURL == "" { // Image URL is required for image2video
				return msgSender.SendMessage(ctx, msgCtx, "Please provide an image URL as the first argument.")
			}
			if prompt == "" {
				// Prompt might be optional for some image2video models, but let's require it for now
				// unless specific models indicate otherwise.
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a text prompt describing the desired animation.")
			}

			// Get model configuration (required for PriceUSD and logic)
			model, exists := faladapter.GetCurrentModel("image2video")
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for image2video"))
			}

			// Don't create client here, use the one in the service
			// client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Video service is now passed in
			// videoService := video.NewVideoService(client, dbManager, bot, debug)

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "image2video", msgCtx.IsPM, msgCtx.GC)

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &video.VideoRequest{
				Prompt:          prompt,
				Duration:        duration,
				AspectRatio:     aspectRatio,
				NegativePrompt:  negativePrompt,
				CFGScale:        cfgScalePtr,        // Assign the parsed pointer
				PromptOptimizer: promptOptimizerPtr, // Assign the parsed pointer (may be nil)
				ModelType:       "image2video",
				Progress:        progress,
				UserNick:        msgCtx.Nick,
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
			result, err := imageService.GenerateVideo(ctx, req)

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, msgCtx, "image2video", result, err); handleErr != nil {
				return handleErr // Propagate error if not handled by the utility function
			}

			// If we reach here, the operation was successful and errors were handled
			return nil
		}),
	}
}
