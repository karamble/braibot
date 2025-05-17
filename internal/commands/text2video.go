package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/video"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Text2VideoCommand returns the text2video command
// It now requires a VideoService instance.
func Text2VideoCommand(bot *kit.Bot, cfg *botconfig.BotConfig, videoService *video.VideoService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("text2video")
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
		Category:    "🎨 AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Create a message sender using the adapter
			msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

			if len(args) < 1 {
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a prompt. Usage: !text2video [prompt] [--duration 5] [--aspect 16:9]")
			}

			// Get the prompt from the arguments
			prompt := strings.Join(args, " ")

			// Create the video request
			req := video.VideoRequest{
				Prompt: prompt,
				IsPM:   msgCtx.IsPM,
				GC:     msgCtx.GC,
			}

			// Process the video
			result, err := videoService.GenerateVideo(ctx, &req)
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, err)
			}

			// Send the result
			return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Generated video: %s", result.VideoURL))
		}),
	}
}
