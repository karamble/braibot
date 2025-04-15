package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2ImageCommand returns the text2image command
func Text2ImageCommand(dbManager *database.DBManager, debug bool) Command {
	return Command{
		Name:        "text2image",
		Description: "Generate an image from text using AI",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				return bot.SendPM(ctx, pm.Nick, "Please provide a prompt for the image generation. Usage: !text2image <prompt>")
			}

			prompt := strings.Join(args, " ")

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("text2image")
			if !exists {
				return fmt.Errorf("no default model found for text2image")
			}
			model, exists := falapi.GetModel(modelName, "text2image")
			if !exists {
				return fmt.Errorf("model not found: %s", modelName)
			}

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
				fmt.Printf("DEBUG - Text2Image command:\n")
				fmt.Printf("  User ID: %s\n", userIDStr)
				fmt.Printf("  Current balance (atoms): %d\n", balance)
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

			// Convert balance to DCR for display
			balanceDCR := float64(balance) / 1e11

			// Send confirmation message with remaining balance
			bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Processing your request. Remaining Balance: %.8f DCR", balanceDCR))

			// Generate image
			imageResp, err := client.GenerateImage(ctx, prompt, model.Name, bot, pm.Nick)
			if err != nil {
				return err
			}

			// Assuming the first image is the one we want to send
			if len(imageResp.Images) > 0 {
				imageURL := imageResp.Images[0].URL
				// Fetch the image data
				imgResp, err := http.Get(imageURL)
				if err != nil {
					return err
				}
				defer imgResp.Body.Close()

				imgData, err := io.ReadAll(imgResp.Body)
				if err != nil {
					return err
				}

				// Encode the image data to base64
				encodedImage := base64.StdEncoding.EncodeToString(imgData)

				// Determine the image type from ContentType
				var imageType string
				switch imageResp.Images[0].ContentType {
				case "image/jpeg":
					imageType = "image/jpeg"
				case "image/png":
					imageType = "image/png"
				case "image/webp":
					imageType = "image/webp"
				default:
					imageType = "image/jpeg" // Fallback to jpeg if unknown
				}

				// Create the message with embedded image, using the user's prompt as the alt text
				message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--", url.QueryEscape(prompt), imageType, encodedImage)
				if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
					return fmt.Errorf("failed to send image: %v", err)
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
				billingMsg := fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
					dcrAmount, model.Price, newBalance)
				if err := bot.SendPM(ctx, pm.Nick, billingMsg); err != nil {
					return fmt.Errorf("failed to send billing information: %v", err)
				}
			} else {
				return bot.SendPM(ctx, pm.Nick, "No images were generated.")
			}

			return nil
		},
	}
}
