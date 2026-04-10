package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/internal/video"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Video2VideoCommand returns the video2video command
func Video2VideoCommand(bot *kit.Bot, cfg *botconfig.BotConfig, videoService *video.VideoService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("video2video", "")
	if !exists {
		model = faladapter.AppModel{Model: fal.Model{
			Name:        "video2video",
			Description: "Edit/transform a video",
		}}
	}
	description := fmt.Sprintf("%s. Usage: !video2video [video_url] [prompt] [--options]", model.Description)

	return braibottypes.Command{
		Name:        "video2video",
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
				model, exists := faladapter.GetCurrentModel("video2video", userIDStr)
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default video2video model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("video2video", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !video2video [video_url] [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			parsed, err := parser.ParseVideo2Video(args)
			if err != nil {
				return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Argument error: %v", err))
			}
			if parsed.Prompt == "" {
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a prompt describing the desired video edit.")
			}

			// Get model configuration
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}
			model, exists := faladapter.GetCurrentModel("video2video", userIDStr)
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for video2video"))
			}

			// Determine effective duration for billing
			duration := parsed.Duration
			durInt := 0
			if duration != "" {
				durInt, err = strconv.Atoi(duration)
				if err != nil || durInt <= 0 {
					durInt = 0
				}
			}
			if durInt == 0 {
				durInt = 5 // Default billing duration
				duration = "5"
			}

			totalCost := model.PriceUSD
			if model.PerSecondPricing {
				totalCost = model.PriceUSD * float64(durInt)
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "video2video", msgCtx.IsPM, msgCtx.GC)

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &video.VideoRequest{
				GenerationRequest: braibottypes.GenerationRequest{
					ModelType: "video2video",
					ModelName: model.Name,
					Progress:  progress,
					UserNick:  msgCtx.Nick,
					UserID:    userID,
					PriceUSD:  totalCost,
					IsPM:      msgCtx.IsPM,
					GC:        msgCtx.GC,
				},
				Prompt:    parsed.Prompt,
				VideoURL:  parsed.VideoURL,
				KeepAudio: parsed.KeepAudio,
				ImageURLs: parsed.ImageURLs,
				Duration:  duration,
			}

			// Inform user of pricing and total cost
			if msgCtx.IsPM {
				if model.PerSecondPricing {
					msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf(
						"Model: %s\n💰 Price: $%.2f per video second\nEstimated duration: %d seconds\nEstimated cost: $%.2f = $%.2f/sec × %d sec",
						model.Name, model.PriceUSD, durInt, totalCost, model.PriceUSD, durInt,
					))
				} else {
					msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf(
						"Model: %s\n💰 Flat fee: $%.2f per video",
						model.Name, model.PriceUSD,
					))
				}
			}

			// Process the video
			result, err := videoService.GenerateVideo(ctx, req)

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, msgCtx, "video2video", result, err); handleErr != nil {
				return handleErr
			}

			return nil
		}),
	}
}
