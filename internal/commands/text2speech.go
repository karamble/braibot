package commands

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/falapi"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(dbManager *database.DBManager, debug bool) Command {
	return Command{
		Name:        "text2speech",
		Description: "Converts text to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 2 {
				voiceList := "Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl"
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Please provide a voice ID and text. Usage: !text2speech [voice_id] [text]\n\n%s", voiceList))
			}

			// Get the text to convert
			text := strings.TrimSpace(args[1])
			if text == "" {
				return fmt.Errorf("please provide text to convert to speech")
			}

			// Get model configuration
			modelName, exists := falapi.GetDefaultModel("text2speech")
			if !exists {
				return fmt.Errorf("no default model found for text2speech")
			}
			model, exists := falapi.GetModel(modelName, "text2speech")
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
				return fmt.Errorf("failed to get balance: %v", err)
			}

			// Check if user has sufficient balance
			if balance < dcrAtoms {
				// Convert balance to DCR for display
				balanceDCR := dcrutil.Amount(balance / 1e3).ToCoin()
				return fmt.Errorf("insufficient balance. Required: %.8f DCR, Current: %.8f DCR", dcrAmount, balanceDCR)
			}

			// Convert balance to DCR for display
			balanceDCR := dcrutil.Amount(balance / 1e3).ToCoin()

			// Send confirmation message with remaining balance
			bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Processing your request. Remaining Balance: %.8f DCR", balanceDCR))

			// Create Fal.ai client
			client := falapi.NewClient(cfg.ExtraConfig["falapikey"], debug)

			// Generate speech
			speechResp, err := client.GenerateSpeech(ctx, text, model.Name, bot, pm.Nick)
			if err != nil {
				return fmt.Errorf("failed to generate speech: %v", err)
			}

			// Log the response for debugging
			if debug {
				fmt.Printf("Speech generation response: %+v\n", speechResp)
			}

			// Check if the audio URL is empty
			if speechResp.Audio.URL == "" {
				return fmt.Errorf("received empty audio URL from API")
			}

			// Fetch the audio data
			audioResp, err := http.Get(speechResp.Audio.URL)
			if err != nil {
				return fmt.Errorf("failed to fetch audio: %v", err)
			}
			defer audioResp.Body.Close()

			audioData, err := io.ReadAll(audioResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read audio data: %v", err)
			}

			// Encode the audio data to base64
			encodedAudio := base64.StdEncoding.EncodeToString(audioData)

			// Create the message with embedded audio
			message := fmt.Sprintf("--embed[alt=%s speech,type=%s,data=%s]--",
				model.Name,
				speechResp.Audio.ContentType,
				encodedAudio)
			if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
				return fmt.Errorf("failed to send audio: %v", err)
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

			// Send billing information with model's USD price and converted DCR amount
			billingMsg := fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• Remaining Balance: %.8f DCR",
				dcrAmount, model.Price, newBalance)
			if err := bot.SendPM(ctx, pm.Nick, billingMsg); err != nil {
				return fmt.Errorf("failed to send billing information: %v", err)
			}

			// Debug information
			if debug {
				fmt.Printf("DEBUG - Text2Speech command:\n")
				fmt.Printf("  User ID: %s\n", userIDStr)
				fmt.Printf("  Current balance (atoms): %d\n", balance)
				fmt.Printf("  Cost in USD: $%.2f\n", model.Price)
				fmt.Printf("  Cost in DCR: %.8f\n", dcrAmount)
				fmt.Printf("  Cost in atoms: %d\n", dcrAtoms)
				fmt.Printf("  Balance in DCR: %.8f\n", float64(balance)/1e11)
			}

			// Debug information after deduction
			if debug {
				fmt.Printf("DEBUG - After deduction:\n")
				fmt.Printf("  New balance in DCR: %.8f\n", newBalance)
			}

			return nil
		},
	}
}
