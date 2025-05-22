package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	braibottypes "github.com/karamble/braibot/internal/types"
	kit "github.com/vctt94/bisonbotkit"
	botconfig "github.com/vctt94/bisonbotkit/config"
)

// WebhookResponse represents the structure of the webhook response
type WebhookResponse struct {
	SessionID         string        `json:"session_id"`
	Query             string        `json:"query"`
	Output            string        `json:"output"`
	IntermediateSteps []interface{} `json:"intermediateSteps"`
}

// AICommand returns the AI command that forwards messages to a webhook
func AICommand(bot *kit.Bot, cfg *botconfig.BotConfig, debug bool) braibottypes.Command {
	return braibottypes.Command{
		Name:        "ai",
		Description: "🤖 Send a message to the AI webhook for processing",
		Category:    "🎨 AI Generation",
		Handler: braibottypes.CommandFunc(func(ctx context.Context, msgCtx braibottypes.MessageContext, args []string, sender *braibottypes.MessageSender, db braibottypes.DBManagerInterface) error {
			// Create a message sender using the adapter
			msgSender := braibottypes.NewMessageSender(braibottypes.NewBisonBotAdapter(bot))

			// Check if webhook is enabled
			webhookEnabled, hasWebhookEnabled := cfg.ExtraConfig["webhookenabled"]
			if !hasWebhookEnabled || webhookEnabled != "true" {
				return msgSender.SendMessage(ctx, msgCtx, "Webhook functionality is not enabled. Try again later.")
			}

			// Check if webhook URL and API key are configured
			webhookURL, hasWebhookURL := cfg.ExtraConfig["webhookurl"]
			webhookAPIKey, hasWebhookAPIKey := cfg.ExtraConfig["webhookapikey"]

			if !hasWebhookURL || !hasWebhookAPIKey {
				return msgSender.SendMessage(ctx, msgCtx, "Webhook not properly configured. Try again later.")
			}

			// Get the full message content
			fullMessage := msgCtx.Message

			// Create request body
			requestBody := map[string]string{
				"message": fullMessage,
				"user":    msgCtx.Nick,
			}
			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to marshal request body: %v", err))
			}

			// Create HTTP client with longer timeout
			client := &http.Client{
				Timeout: 120 * time.Second, // 120 second timeout (2 minutes)
			}

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
			defer cancel()

			// Create request with context
			req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to create request: %v", err))
			}

			// Set headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-BRAIBOT-API-KEY", webhookAPIKey)

			if debug {
				fmt.Printf("DEBUG [ai] User %s: Sending request to webhook\n", msgCtx.Nick)
			}

			// Send request
			resp, err := client.Do(req)
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to send request to webhook: %v", err))
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to read response body: %v", err))
			}

			// Debug: Log the raw response
			if debug {
				fmt.Printf("DEBUG [ai] User %s: Webhook response body: %s\n", msgCtx.Nick, string(body))
			}

			// Check response status
			if resp.StatusCode != http.StatusOK {
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("webhook returned error status %d: %s", resp.StatusCode, string(body)))
			}

			// Parse response as array of WebhookResponse
			var responses []WebhookResponse
			if err := json.Unmarshal(body, &responses); err != nil {
				if debug {
					fmt.Printf("DEBUG [ai] User %s: Failed to parse response as JSON: %v\n", msgCtx.Nick, err)
				}
				return msgSender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to parse response as JSON: %v", err))
			}

			// Debug: Log the parsed responses
			if debug {
				fmt.Printf("DEBUG [ai] User %s: Number of responses: %d\n", msgCtx.Nick, len(responses))
			}

			// Check if we have at least one response
			if len(responses) == 0 {
				return msgSender.SendMessage(ctx, msgCtx, "Unable to process your query: no response received.")
			}

			// Handle different response formats
			var output string
			var sessionID string
			if len(responses) == 2 {
				// Voice command format: second response contains the output
				output = responses[1].Output
				sessionID = responses[0].SessionID
			} else {
				// Text command format: first response contains the output
				output = responses[0].Output
				sessionID = responses[0].SessionID
			}

			// Validate output
			if output == "" {
				if debug {
					fmt.Printf("DEBUG [ai] User %s: Missing output in response\n", msgCtx.Nick)
				}
				return msgSender.SendMessage(ctx, msgCtx, "Unable to process your query: no output received.")
			}

			// Validate session_id
			if sessionID == "" {
				if debug {
					fmt.Printf("DEBUG [ai] User %s: Missing session_id in response\n", msgCtx.Nick)
				}
				// Fallback to original nick if session_id is missing
				sessionID = msgCtx.Nick
			}

			if debug {
				fmt.Printf("DEBUG [ai] User %s: Sending response output to session %s\n", msgCtx.Nick, sessionID)
			}

			// Send only the output field back to the appropriate channel based on the original message context
			if msgCtx.IsPM {
				return bot.SendPM(ctx, sessionID, output)
			} else {
				return bot.SendGC(ctx, msgCtx.GC, output)
			}
		}),
	}
}
