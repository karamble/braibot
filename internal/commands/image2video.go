package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2VideoCommand returns the image2video command
func Image2VideoCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2video")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2video",
			Description: "Convert an image to video using AI",
		}
	}

	// Create the command description using the model's description
	description := fmt.Sprintf("%s. Usage: !image2video [image_url] [prompt] [--model name] [--duration seconds] [--aspect ratio] [--negative-prompt text] [--cfg-scale value]", model.Description)

	return Command{
		Name:        "image2video",
		Description: description,
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 2 {
				// Get the current model to use its help documentation
				model, exists := faladapter.GetCurrentModel("image2video")
				if !exists {
					return bot.SendPM(ctx, pm.Nick, "Please provide an image URL and prompt. Usage: !image2video [image_url] [prompt] [--model name] [--duration seconds] [--aspect ratio] [--negative-prompt text] [--cfg-scale value]")
				}

				// Use the model's help documentation if available
				if model.HelpDoc != "" {
					return bot.SendPM(ctx, pm.Nick, model.HelpDoc)
				}

				// Fallback to default help message
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL and prompt. Usage: !image2video [image_url] [prompt] [--model name] [--duration seconds] [--aspect ratio] [--negative-prompt text] [--cfg-scale value]")
			}

			// Parse arguments
			imageURL := args[0]

			// Validate image URL
			if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") && !strings.HasPrefix(imageURL, "data:") {
				// Try to add https:// if no scheme is provided
				imageURL = "https://" + imageURL
			}

			var prompt string
			var modelName string
			var duration string = "5s"                                   // Default duration for Veo2
			var aspectRatio string = "auto"                              // Default aspect ratio for Veo2
			var negativePrompt string = "blur, distort, and low quality" // Default negative prompt for Kling
			var cfgScale float64 = 0.5                                   // Default CFG scale for Kling

			// Process arguments
			for i := 1; i < len(args); i++ {
				if args[i] == "--model" && i+1 < len(args) {
					modelName = args[i+1]
					i++ // Skip the next argument
				} else if args[i] == "--duration" && i+1 < len(args) {
					duration = args[i+1]
					i++ // Skip the next argument
				} else if args[i] == "--aspect" && i+1 < len(args) {
					aspectRatio = args[i+1]
					i++ // Skip the next argument
				} else if args[i] == "--negative-prompt" && i+1 < len(args) {
					negativePrompt = args[i+1]
					i++ // Skip the next argument
				} else if args[i] == "--cfg-scale" && i+1 < len(args) {
					scale, err := strconv.ParseFloat(args[i+1], 64)
					if err == nil {
						cfgScale = scale
						i++ // Skip the next argument
					}
				} else {
					// If not a flag, add to prompt
					if prompt == "" {
						prompt = args[i]
					} else {
						prompt += " " + args[i]
					}
				}
			}

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("image2video")
			if !exists {
				return fmt.Errorf("no default model found for image2video")
			}

			// If model name is provided, override the default
			if modelName != "" {
				newModel, exists := faladapter.GetModel(modelName, "image2video")
				if !exists {
					return fmt.Errorf("model not found: %s", modelName)
				}
				model = newModel
			}

			// Calculate price based on model
			var totalPrice float64
			if model.Name == "kling-video" {
				// Parse duration for Kling
				dur, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
				if err != nil {
					dur = 5 // Default to 5 seconds if parsing fails
				}
				basePrice := model.PriceUSD // Base price for 5 seconds
				additionalSeconds := 0
				if dur > 5 {
					additionalSeconds = dur - 5
				}
				totalPrice = basePrice + (float64(additionalSeconds) * 0.4)
			} else if model.Name == "veo2" {
				// Parse duration for Veo2
				dur, err := strconv.Atoi(strings.TrimSuffix(duration, "s"))
				if err != nil {
					dur = 5 // Default to 5 seconds if parsing fails
				}
				basePrice := model.PriceUSD // Base price for 5 seconds
				additionalSeconds := 0
				if dur > 5 {
					additionalSeconds = dur - 5
				}
				totalPrice = basePrice + (float64(additionalSeconds) * 0.50)
			} else {
				// Fallback to model's base price
				totalPrice = model.PriceUSD
			}

			// Override the model price with the calculated price
			model.PriceUSD = totalPrice

			// Convert model's USD price to DCR using current exchange rate
			dcrAmount, err := USDToDCR(model.PriceUSD)
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error: %v", err))
			}

			// Convert DCR amount to atoms for comparison (1 DCR = 1e11 atoms)
			dcrAtoms := int64(dcrAmount * 1e11)

			// Get user balance in atoms
			var userID zkidentity.ShortID
			userID.FromBytes(pm.Uid)
			userIDStr := userID.String()
			balance, err := dbManager.GetBalance(userIDStr)
			if err != nil {
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error: %v", err))
			}

			// Check if user has sufficient balance
			if balance < dcrAtoms {
				// Convert balance to DCR for display
				balanceDCR := float64(balance) / 1e11
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Insufficient balance. Required: %.8f DCR, Current: %.8f DCR", dcrAmount, balanceDCR))
			}

			// Send confirmation message
			bot.SendPM(ctx, pm.Nick, "Processing your request.")

			// Create progress callback
			progress := faladapter.NewBotProgressCallback(bot, pm.Nick)

			// Create video request
			req := fal.VideoRequest{
				Prompt:         prompt,
				ImageURL:       imageURL,
				Model:          model.Name,
				Duration:       duration,
				AspectRatio:    aspectRatio,
				NegativePrompt: negativePrompt,
				CFGScale:       cfgScale,
				Progress:       progress,
				Options:        make(map[string]interface{}),
			}

			// Generate video
			resp, err := client.GenerateVideo(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate video: %v", err)
			}

			// Get the video URL from any of the possible fields
			videoURL := resp.GetURL()

			// Send the video URL to the user immediately
			if err := bot.SendPM(ctx, pm.Nick, fmt.Sprintf("🎥 Your video is ready! Here's the direct URL: %s", videoURL)); err != nil {
				return fmt.Errorf("failed to send video URL: %v", err)
			}

			// Try to fetch and embed the video
			videoResp, err := http.Get(videoURL)
			if err != nil {
				// Log the error but don't return it since we already sent the URL
				fmt.Printf("Warning: failed to fetch video for embedding: %v\n", err)
			} else {
				defer videoResp.Body.Close()

				videoData, err := io.ReadAll(videoResp.Body)
				if err != nil {
					fmt.Printf("Warning: failed to read video data for embedding: %v\n", err)
				} else {
					// Encode the video data to base64
					encodedVideo := base64.StdEncoding.EncodeToString(videoData)

					// Create the message with embedded video
					message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--",
						"Generated video",
						"video/mp4", // Default content type for MP4 videos
						encodedVideo)
					if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
						fmt.Printf("Warning: failed to send embedded video: %v\n", err)
					}
				}
			}

			// Deduct balance using CheckAndDeductBalance after successful delivery
			hasBalance, err := dbManager.CheckAndDeductBalance(pm.Uid, model.PriceUSD, debug)
			if err != nil {
				return fmt.Errorf("failed to deduct balance: %v", err)
			}
			if !hasBalance {
				return fmt.Errorf("failed to deduct balance. Please try again.")
			}

			// Get updated balance for billing message
			newBalance, err := dbManager.GetUserBalance(pm.Uid)
			if err != nil {
				return fmt.Errorf("failed to get updated balance: %v", err)
			}

			// Send billing information with model's USD price and converted DCR amount
			billingMsg := fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
				dcrAmount, model.PriceUSD, newBalance)
			if err := bot.SendPM(ctx, pm.Nick, billingMsg); err != nil {
				return fmt.Errorf("failed to send billing information: %v", err)
			}

			return nil
		},
	}
}
