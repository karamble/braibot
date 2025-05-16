package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// WebhookResponse represents the structure of the webhook response
type WebhookResponse struct {
	SessionID         string        `json:"session_id"`
	Query             string        `json:"query"`
	Output            string        `json:"output"`
	IntermediateSteps []interface{} `json:"intermediateSteps"`
}

// AICommand returns the AI command that forwards messages to a webhook
func AICommand() Command {
	return Command{
		Name:        "ai",
		Description: "🤖 Send a message to the AI webhook for processing",
		Category:    "🎨 AI Generation",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			// Check if webhook is enabled
			webhookEnabled, hasWebhookEnabled := cfg.ExtraConfig["webhookenabled"]
			if !hasWebhookEnabled || webhookEnabled != "true" {
				return bot.SendPM(ctx, pm.Nick, "Webhook functionality is not enabled. Try again later.")
			}

			// Check if webhook URL and API key are configured
			webhookURL, hasWebhookURL := cfg.ExtraConfig["webhookurl"]
			webhookAPIKey, hasWebhookAPIKey := cfg.ExtraConfig["webhookapikey"]

			if !hasWebhookURL || !hasWebhookAPIKey {
				return bot.SendPM(ctx, pm.Nick, "Webhook not properly configured. Try again later.")
			}

			// Get the full message content
			fullMessage := pm.Msg.Message

			// Create request body
			requestBody := map[string]string{
				"message": fullMessage,
				"user":    pm.Nick,
			}
			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %v", err)
			}

			// Create HTTP client with longer timeout
			client := &http.Client{
				Timeout: 60 * time.Second, // 60 second timeout
			}

			// Create request
			req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				return fmt.Errorf("failed to create request: %v", err)
			}

			// Set headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-BRAIBOT-API-KEY", webhookAPIKey)

			// Send request
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("failed to send request to webhook: %v", err)
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %v", err)
			}

			// Debug: Log the raw response
			log.Printf("Webhook response body: %s", string(body))

			// Check response status
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("webhook returned error status %d: %s", resp.StatusCode, string(body))
			}

			// Parse response as array of WebhookResponse
			var responses []WebhookResponse
			if err := json.Unmarshal(body, &responses); err != nil {
				log.Printf("Failed to parse JSON: %v", err)
				return fmt.Errorf("failed to parse response as JSON: %v", err)
			}

			// Debug: Log the parsed responses
			log.Printf("Number of responses: %d", len(responses))
			if len(responses) > 0 {
				log.Printf("First response - SessionID: %q", responses[0].SessionID)
			}
			if len(responses) > 1 {
				log.Printf("Second response - Output: %q", responses[1].Output)
			}

			// Check if we have at least two responses (query and output)
			if len(responses) < 2 {
				log.Printf("Not enough responses: got %d, expected at least 2", len(responses))
				return bot.SendPM(ctx, pm.Nick, "Unable to process your query.")
			}

			// Get session_id from first response and output from second response
			sessionID := responses[0].SessionID
			output := responses[1].Output

			// Validate required fields
			if sessionID == "" || output == "" {
				log.Printf("Missing required fields - SessionID: %q, Output: %q", sessionID, output)
				return bot.SendPM(ctx, pm.Nick, "Unable to process your query.")
			}

			// Send the output back to the user identified by session_id
			return bot.SendPM(ctx, sessionID, output)
		},
	}
}
