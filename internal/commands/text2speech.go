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
	"github.com/karamble/braibot/internal/audio"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/faladapter"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Text2SpeechCommand returns the text2speech command
func Text2SpeechCommand(dbManager *database.DBManager, debug bool) Command {
	return Command{
		Name:        "text2speech",
		Description: "Converts text to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			if len(args) < 1 {
				voiceList := "Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl"
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Please provide text to convert to speech. Usage: !text2speech [voice_id] [text]\n\n%s", voiceList))
			}

			// Get the text to convert - join all arguments after the voice ID
			var text string
			var voiceID string

			if len(args) > 1 {
				// If voice ID is provided, use it and join remaining args for text
				voiceID = args[0]
				text = strings.Join(args[1:], " ")
			} else {
				// If no voice ID provided, use default and use all args for text
				voiceID = "Wise_Woman" // Default voice ID
				text = strings.Join(args, " ")
			}

			if text == "" {
				return fmt.Errorf("please provide text to convert to speech")
			}

			// Get model configuration
			model, exists := faladapter.GetCurrentModel("text2speech")
			if !exists {
				return fmt.Errorf("no default model found for text2speech")
			}

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
				return fmt.Errorf("failed to get balance: %v", err)
			}

			// Check if user has sufficient balance
			if balance < dcrAtoms {
				// Convert balance to DCR for display
				balanceDCR := dcrutil.Amount(balance / 1e3).ToCoin()
				return fmt.Errorf("insufficient balance. Required: %.8f DCR, Current: %.8f DCR", dcrAmount, balanceDCR)
			}

			// Send confirmation message
			bot.SendPM(ctx, pm.Nick, "Processing your request.")

			// Create Fal.ai client
			client := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

			// Create progress callback
			progress := NewCommandProgressCallback(bot, pm.Nick, "text2speech")

			// Create speech request
			req := fal.SpeechRequest{
				Text:     text,
				VoiceID:  model.Name, // Use the model name from configuration
				Progress: progress,
				Options: map[string]interface{}{
					"voice_id": voiceID, // Pass the voice ID as an option
				},
			}

			// Generate speech
			resp, err := client.GenerateSpeech(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to generate speech: %v", err)
			}

			// Log the response for debugging
			if debug {
				fmt.Printf("Speech generation response: %+v\n", resp)
			}

			// Check if the audio URL is empty
			if resp.AudioURL == "" {
				return fmt.Errorf("received empty audio URL from API")
			}

			// Fetch the audio data
			audioResp, err := http.Get(resp.AudioURL)
			if err != nil {
				return fmt.Errorf("failed to fetch audio: %v", err)
			}
			defer audioResp.Body.Close()

			audioData, err := io.ReadAll(audioResp.Body)
			if err != nil {
				return fmt.Errorf("failed to read audio data: %v", err)
			}

			// Convert the audio data to OGG/Opus format
			opusData, err := audio.ConvertPCMToOpus(audioData)
			if err != nil {
				return fmt.Errorf("failed to convert audio to Opus: %v", err)
			}

			// Encode the opus data to base64
			encodedAudio := base64.StdEncoding.EncodeToString(opusData)

			// Create the message with embedded audio
			message := fmt.Sprintf("--embed[alt=%s speech,type=audio/ogg,data=%s]--",
				model.Name,
				encodedAudio)
			if err := bot.SendPM(ctx, pm.Nick, message); err != nil {
				return fmt.Errorf("failed to send audio: %v", err)
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

			// Debug information
			if debug {
				fmt.Printf("DEBUG - Text2Speech command:\n")
				fmt.Printf("  User ID: %s\n", userIDStr)
				fmt.Printf("  Current balance (atoms): %d\n", balance)
				fmt.Printf("  Cost in USD: $%.2f\n", model.PriceUSD)
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
