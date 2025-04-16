package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2VideoCommand returns the image2video command
func Image2VideoCommand(dbManager *database.DBManager, debug bool) Command {
	return Command{
		Name:        "image2video",
		Description: "Convert an image to video using AI. Usage: !image2video [image_url] [prompt] [--duration seconds] [--aspect ratio] [--negative-prompt text] [--cfg-scale value]",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 2 {
				return bot.SendPM(ctx, pm.Nick, "Please provide an image URL and prompt. Usage: !image2video [image_url] [prompt] [--duration seconds] [--aspect ratio] [--negative-prompt text] [--cfg-scale value]")
			}

			// Parse arguments
			imageURL := args[0]
			var prompt string
			var duration int = 5                                         // Default duration
			var aspectRatio string = "16:9"                              // Default aspect ratio
			var negativePrompt string = "blur, distort, and low quality" // Default negative prompt
			var cfgScale float64 = 0.5                                   // Default CFG scale

			// Process arguments
			for i := 1; i < len(args); i++ {
				if args[i] == "--duration" && i+1 < len(args) {
					dur, err := strconv.Atoi(args[i+1])
					if err == nil && dur >= 5 {
						duration = dur
						i++ // Skip the next argument
					}
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

			// Calculate price based on duration
			basePrice := 2.0 // Base price for 5 seconds
			additionalSeconds := 0
			if duration > 5 {
				additionalSeconds = duration - 5
			}
			totalPrice := basePrice + (float64(additionalSeconds) * 0.4)

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("image2video")
			if !exists {
				return fmt.Errorf("no default model found for image2video")
			}
			model, exists := falapi.GetModel(modelName, "image2video")
			if !exists {
				return fmt.Errorf("model not found: %s", modelName)
			}

			// Override the model price with the calculated price
			model.Price = totalPrice

			// Convert model's USD price to DCR using current exchange rate
			dcrAmount, err := USDToDCR(model.Price)
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

			// Debug information
			if debug {
				fmt.Printf("DEBUG - Image2Video command:\n")
				fmt.Printf("  User ID: %s\n", userIDStr)
				fmt.Printf("  Current balance (atoms): %d\n", balance)
				fmt.Printf("  Duration: %d seconds\n", duration)
				fmt.Printf("  Cost in USD: $%.2f\n", model.Price)
				fmt.Printf("  Cost in DCR: %.8f\n", dcrAmount)
				fmt.Printf("  Cost in atoms: %d\n", dcrAtoms)
				fmt.Printf("  Balance in DCR: %.8f\n", float64(balance)/1e11)
			}

			// Check if user has sufficient balance
			if balance < dcrAtoms {
				// Convert balance to DCR for display
				balanceDCR := float64(balance) / 1e11
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Insufficient balance. You have %.8f DCR, but this operation requires %.8f DCR (%.2f USD). Please send a tip to use this feature.",
					balanceDCR, dcrAmount, model.Price))
			}

			// Send confirmation message
			bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Processing your request. The video generation will take a while. The bot does not respond to messages while processing. The process can take up to 20 minutes. Please be patient. Video Duration: %d seconds.", duration))

			// Generate video
			videoResp, err := client.GenerateVideoFromImage(ctx, prompt, imageURL, model.Name, bot, pm.Nick, duration, aspectRatio, negativePrompt, cfgScale)
			if err != nil {
				return fmt.Errorf("failed to generate video: %v", err)
			}

			// Log the response for debugging
			if debug {
				fmt.Printf("Video generation response: %+v\n", videoResp)
			}

			// Check if the video URL is empty
			if videoResp.Video.URL == "" {
				return fmt.Errorf("received empty video URL from API")
			}

			// Send the video URL to the user
			if err := bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Your video is ready: %s", videoResp.Video.URL)); err != nil {
				return fmt.Errorf("failed to send video URL: %v", err)
			}

			// Fetch the video data
			videoHTTPResp, err := http.Get(videoResp.Video.URL)
			if err != nil {
				return fmt.Errorf("failed to fetch video: %v", err)
			}
			defer videoHTTPResp.Body.Close()

			videoData, err := io.ReadAll(videoHTTPResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read video data: %v", err)
			}

			// Encode the video data to base64
			encodedVideo := base64.StdEncoding.EncodeToString(videoData)

			// Create the message with embedded video
			message := fmt.Sprintf("--embed[alt=%s video,type=%s,data=%s]--",
				model.Name,
				videoResp.Video.ContentType,
				encodedVideo)
			if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
				return fmt.Errorf("failed to send video: %v", err)
			}

			// Deduct balance using CheckAndDeductBalance after successful delivery
			hasBalance, err := dbManager.CheckAndDeductBalance(pm.Uid, model.Price, debug)
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

			// Debug information after deduction
			if debug {
				fmt.Printf("DEBUG - After deduction:\n")
				fmt.Printf("  New balance in DCR: %.8f\n", newBalance)
			}

			// Send billing information with model's USD price and converted DCR amount
			billingMsg := fmt.Sprintf("💰 Billing Information:\n• Duration: %d seconds\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
				duration, dcrAmount, model.Price, newBalance)
			if err := bot.SendPM(ctx, pm.Nick, billingMsg); err != nil {
				return fmt.Errorf("failed to send billing information: %v", err)
			}

			return nil
		},
	}
}
