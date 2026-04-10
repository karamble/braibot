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

// Multi2VideoCommand returns the multi2video (reference-to-video) command.
// It accepts a prompt plus any combination of reference images (up to 9),
// videos (up to 3), and audio files (up to 3).
func Multi2VideoCommand(bot *kit.Bot, cfg *botconfig.BotConfig, videoService *video.VideoService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("multi2video", "")
	if !exists {
		model = fal.Model{
			Name:        "multi2video",
			Description: "Generate a video from a prompt plus reference images, videos, and audio",
		}
	}
	description := fmt.Sprintf("%s. Usage: !multi2video [prompt] [--image1..9 url] [--video1..3 url] [--audio1..3 url] [--options]", model.Description)

	return braibottypes.Command{
		Name:        "multi2video",
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
				model, exists := faladapter.GetCurrentModel("multi2video", userIDStr)
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default multi2video model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("multi2video", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !multi2video [prompt] [--image1..9 url] [--video1..3 url] [--audio1..3 url] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			// Parse arguments using the video parser
			parser := video.NewArgumentParser()
			parsed, err := parser.ParseMulti2Video(args)
			if err != nil {
				return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Argument error: %v", err))
			}
			if parsed.Prompt == "" {
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a text prompt describing the desired video.")
			}
			if len(parsed.ImageURLs) == 0 && len(parsed.VideoURLs) == 0 && len(parsed.AudioURLs) == 0 {
				return msgSender.SendMessage(ctx, msgCtx, "At least one reference input (--image1, --video1, or --audio1) is required. For a purely text-driven video, use !text2video instead.")
			}
			if len(parsed.AudioURLs) > 0 && len(parsed.ImageURLs) == 0 && len(parsed.VideoURLs) == 0 {
				return msgSender.SendMessage(ctx, msgCtx, "Reference audio requires at least one reference image or video.")
			}

			// Get model configuration
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}
			model, exists := faladapter.GetCurrentModel("multi2video", userIDStr)
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for multi2video"))
			}

			// Determine effective duration for per-second pricing
			duration := parsed.Duration
			originalUserDuration := duration
			durInt := 0
			if duration != "" {
				durInt, err = strconv.Atoi(duration)
				if err != nil || durInt <= 0 {
					durInt = 0
				}
			}
			if durInt == 0 {
				modelDefault := 0
				switch model.Name {
				case "seedance-2.0-reference":
					modelDefault = 5
				}
				if modelDefault > 0 {
					durInt = modelDefault
					duration = strconv.Itoa(modelDefault)
				} else {
					durInt = 5
					duration = "5"
				}
			}

			totalCost := model.PriceUSD
			if model.PerSecondPricing {
				totalCost = model.PriceUSD * float64(durInt)
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "multi2video", msgCtx.IsPM, msgCtx.GC)

			// Create video request using parsed values
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &video.VideoRequest{
				GenerationRequest: braibottypes.GenerationRequest{
					ModelType: "multi2video",
					ModelName: model.Name,
					Progress:  progress,
					UserNick:  msgCtx.Nick,
					UserID:    userID,
					PriceUSD:  totalCost,
					IsPM:      msgCtx.IsPM,
					GC:        msgCtx.GC,
				},
				Prompt:        parsed.Prompt,
				Duration:      duration,
				AspectRatio:   parsed.AspectRatio,
				Resolution:    parsed.Resolution,
				GenerateAudio: parsed.GenerateAudio,
				ImageURLs:     parsed.ImageURLs,
				VideoURLs:     parsed.VideoURLs,
				AudioURLs:     parsed.AudioURLs,
				Seed:          parsed.Seed,
			}

			// Inform user of pricing and total cost
			if msgCtx.IsPM {
				if model.PerSecondPricing {
					msg := fmt.Sprintf(
						"Model: %s\n💰 Price: $%.2f per video second\nRequested duration: %d seconds\nTotal cost: $%.2f = $%.2f/sec × %d sec\nReference inputs: %d image(s), %d video(s), %d audio(s)",
						model.Name, model.PriceUSD, durInt, totalCost, model.PriceUSD, durInt,
						len(parsed.ImageURLs), len(parsed.VideoURLs), len(parsed.AudioURLs),
					)
					if originalUserDuration == "" {
						msg += fmt.Sprintf("\n(No duration specified, using default duration of %d seconds.)", durInt)
					}
					msgSender.SendMessage(ctx, msgCtx, msg)
				} else {
					msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf(
						"Model: %s\n💰 Flat fee: $%.2f per video",
						model.Name, model.PriceUSD,
					))
				}
			}

			// Generate video using the service
			result, err := videoService.GenerateVideo(ctx, req)

			// Handle result/error using the utility function
			if handleErr := utils.HandleServiceResultOrError(ctx, bot, msgCtx, "multi2video", result, err); handleErr != nil {
				return handleErr
			}

			return nil
		}),
	}
}
