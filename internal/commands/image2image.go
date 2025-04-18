package commands

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/internal/image"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2ImageCommand returns the image2image command
// It now requires an ImageService instance.
func Image2ImageCommand(dbManager *database.DBManager, imageService *image.ImageService, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2image")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2image",
			Description: "Transform an image using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !image2image [image_url]", model.Description)

	return Command{
		Name:        "image2image",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("image2image")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2image [image_url]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL. Usage: !image2image [image_url]")
			}

			imageURL := args[0]

			// Validate URL
			parsedURL, err := url.Parse(imageURL)
			if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
				return bot.SendPM(ctx, pm.Nick, "Please provide a valid http:// or https:// URL for the image.")
			}

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("image2image")
			if !exists {
				return fmt.Errorf("no default model found for image2image")
			}

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "image2image")

			// Create image request
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			req := &image.ImageRequest{
				ImageURL:  imageURL,
				ModelType: "image2image",
				ModelName: model.Name,
				Progress:  progress,
				UserNick:  pm.Nick,
				UserID:    userID,
				PriceUSD:  model.PriceUSD,
			}

			// Generate image using the service
			result, err := imageService.GenerateImage(ctx, req)
			if err != nil {
				var insufficientBalanceErr *utils.ErrInsufficientBalance // Define variable outside switch
				switch {
				case errors.As(err, &insufficientBalanceErr):
					// Send specific PM ONLY for insufficient balance
					pmMsg := fmt.Sprintf("Image generation failed: %s", insufficientBalanceErr.Error())
					_ = bot.SendPM(ctx, pm.Nick, pmMsg)
					return nil // Return nil as we notified the user
				case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
					// Context was cancelled (likely due to shutdown signal), log and return nil
					fmt.Printf("INFO [image2image] User %s: Context canceled/deadline exceeded: %v\n", pm.Nick, err)
					return nil // Indicate clean termination due to context cancellation
				default:
					// For ALL other errors, log and return the error to the framework
					fmt.Printf("ERROR [image2image] User %s: %v\n", pm.Nick, err)
					return err // Return the original error
				}
			}

			if !result.Success {
				// Log the error and return it.
				errMsg := fmt.Sprintf("ERROR [image2image] User %s: Image generation failed internally", pm.Nick)
				if result.Error != nil {
					errMsg += fmt.Sprintf(": %v", result.Error)
				}
				fmt.Println(errMsg)
				// Return an error to the framework
				if result.Error != nil {
					return fmt.Errorf("image generation failed: %w", result.Error)
				}
				return fmt.Errorf("image generation failed internally")
			}

			return nil
		},
	}
}
