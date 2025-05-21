package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/faladapter"
	imgservice "github.com/karamble/braibot/internal/image"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// Image2ImageCommand returns the image2image command
// It now requires an ImageService instance.
func Image2ImageCommand(bot *kit.Bot, cfg *botconfig.BotConfig, imageService *imgservice.ImageService, debug bool) braibottypes.Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2image", "") // Empty string for global default
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2image",
			Description: "Transform an image using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !image2image [image_url]", model.Description)

	return braibottypes.Command{
		Name:        "image2image",
		Description: description,
		Category:    "🎨 AI Generation",
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
				model, exists := faladapter.GetCurrentModel("image2image", userIDStr)
				if !exists {
					return msgSender.SendMessage(ctx, msgCtx, "Error: Default image2image model not found.")
				}

				// Get user ID
				var userID zkidentity.ShortID
				userID.FromBytes(msgCtx.Uid)

				// Format header using utility function
				header := utils.FormatCommandHelpHeader("image2image", model, userID, db)

				// Get help doc
				helpDoc := model.HelpDoc
				if helpDoc == "" {
					helpDoc = "Usage: !image2image [image_url] [prompt] [--options...]\n(No specific documentation available for this model.)"
				}

				// Send combined header and help doc
				return msgSender.SendMessage(ctx, msgCtx, header+helpDoc)
			}

			imageURL := args[0]

			// Validate URL
			parsedURL, err := url.Parse(imageURL)
			if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				return msgSender.SendMessage(ctx, msgCtx, "Please provide a valid http:// or https:// URL for the image.")
			}

			// Get model configuration
			var userIDStr string
			if msgCtx.IsPM {
				var uid zkidentity.ShortID
				uid.FromBytes(msgCtx.Uid)
				userIDStr = uid.String()
			}
			model, exists := faladapter.GetCurrentModel("image2image", userIDStr)
			if !exists {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("no default model found for image2image"))
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, msgCtx.Nick, msgCtx.Sender, "image2image", msgCtx.IsPM, msgCtx.GC)

			// Create image request
			var userID zkidentity.ShortID
			userID.FromBytes(msgCtx.Uid)
			req := &imgservice.ImageRequest{
				ImageURL:  imageURL,
				ModelType: "image2image",
				ModelName: model.Name,
				Progress:  progress,
				UserNick:  msgCtx.Nick,
				UserID:    userID,
				PriceUSD:  model.PriceUSD,
			}

			// Generate image using the service
			result, err := imageService.GenerateImage(ctx, req)
			if err != nil {
				var insufficientBalanceErr *utils.ErrInsufficientBalance // Define variable outside switch
				switch {
				case errors.As(err, &insufficientBalanceErr):
					// Send specific message for insufficient balance
					return msgSender.SendMessage(ctx, msgCtx, fmt.Sprintf("Image generation failed: %s", insufficientBalanceErr.Error()))
				case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
					// Context was cancelled (likely due to shutdown signal), log and return nil
					fmt.Printf("INFO [image2image] User %s: Context canceled/deadline exceeded: %v\n", msgCtx.Nick, err)
					return nil // Indicate clean termination due to context cancellation
				default:
					// For ALL other errors, log and return the error to the framework
					fmt.Printf("ERROR [image2image] User %s: %v\n", msgCtx.Nick, err)
					return err // Return the original error
				}
			}

			if !result.Success {
				// Log the error and return it.
				errMsg := fmt.Sprintf("ERROR [image2image] User %s: Image generation failed internally", msgCtx.Nick)
				if result.Error != nil {
					errMsg += fmt.Sprintf(": %v", result.Error)
				}
				fmt.Println(errMsg)
				// Return an error to the framework
				if result.Error != nil {
					return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("image generation failed: %w", result.Error))
				}
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("image generation failed internally"))
			}

			return nil
		}),
	}
}
