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

// Text2VideoCommand returns the text2video command
// It now requires a VideoService instance.
func Text2VideoCommand(bot *kit.Bot, cfg *botconfig.BotConfig, videoService *video.VideoService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2video", "") // Empty string for global default
	if !exists {
		model = fal.Model{
			Name:        "text2video",
			Description: "Generate a video from text",
		}
	}
	description := fmt.Sprintf("%s. Usage: !text2video [prompt] [--duration 5] [--aspect 16:9]", model.Description)

	return braibottypes.Command{
		Name:        "text2video",
		Description: description,
		Category:    "AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Create a message sender using the adapter
			msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

			if len(args) < 1 {
				// Get the current model
				var userIDStr string
				if msgCtx.IsPM {
					var uid zkidentity.ShortID
					uid.FromBytes(msgCtx.Uid)
					userIDStr = uid.String()
				}
				model, exists := faladapter.GetCurrentModel("text2video", userIDStr)
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default text2video model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("text2video", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !text2video [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			prompt, _, duration, aspectRatio, negativePrompt, cfgScalePtr, promptOptimizerPtr, err := parser.Parse(args, false) // No Image URL expected
			if err != nil {
				return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Argument error: %v", err))
			}
			if prompt == "" {
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a text prompt describing the desired video.")
			}

			// Get model configuration
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}
			model, exists := faladapter.GetCurrentModel("text2video", userIDStr)
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for text2video"))
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "text2video", msgCtx.IsPM, msgCtx.GC)

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &video.VideoRequest{
				Prompt:          prompt,
				Duration:        duration,
				AspectRatio:     aspectRatio,
				NegativePrompt:  negativePrompt,
				CFGScale:        cfgScalePtr,
				PromptOptimizer: promptOptimizerPtr,
				ModelType:       "text2video",
				Progress:        progress,
				UserNick:        msgCtx.Nick,
				UserID:          userID,
				PriceUSD:        model.PriceUSD,
			}

			// Process the video
			result, err := videoService.GenerateVideo(ctx, req)
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, err)
			}

			// Send the result
			return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Generated video: %s", result.VideoURL))
		}),
	}
}
