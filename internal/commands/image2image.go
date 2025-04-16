package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Image2ImageCommand returns the image2image command
func Image2ImageCommand(dbManager *database.DBManager, debug bool) Command {
	// Get the current model to use its description
	model, exists := faladapter.GetCurrentModel("image2image")
	if !exists {
		// Fallback to a default description if no model is found
		model = fal.Model{
			Name:        "image2image",
			Description: "Transforms an image using AI",
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

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

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

			// Create image request
			req := fal.ImageRequest{
				Model: model.Name,
				Options: map[string]interface{}{
					"image_url": imageURL,
				},
				Progress: progress,
			}

			// Generate image
			resp, err := client.GenerateImage(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate image: %v", err)
			}

			// Log the response for debugging
			if debug {
				fmt.Printf("Image generation response: %+v\n", resp)
			}

			// Check if the image URL is empty
			if len(resp.Images) == 0 || resp.Images[0].URL == "" {
				return fmt.Errorf("received empty image URL from API")
			}

			// Fetch the image data
			imgResp, err := http.Get(resp.Images[0].URL)
			if err != nil {
				return fmt.Errorf("failed to fetch image: %v", err)
			}
			defer imgResp.Body.Close()

			imgData, err := io.ReadAll(imgResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read image data: %v", err)
			}

			// Encode the image data to base64
			encodedImage := base64.StdEncoding.EncodeToString(imgData)

			// Create the message with embedded image
			message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--",
				url.QueryEscape("Transformed image"),
				resp.Images[0].ContentType,
				encodedImage)
			if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
				return fmt.Errorf("failed to send image: %v", err)
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
