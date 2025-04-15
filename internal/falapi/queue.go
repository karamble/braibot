// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

// pollQueueStatus polls the queue status until completion or error
func (c *Client) pollQueueStatus(ctx context.Context, queueResp QueueResponse, bot interface{}, userNick string) (*QueueResponse, error) {
	// Create a ticker for polling
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Poll status until completion
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Get status using status_url
			statusURL := queueResp.StatusURL + "?logs=1"
			req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create status request: %v", err)
			}
			req.Header.Set("Authorization", "Key "+c.apiKey)

			statusResp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %v", err)
			}
			defer statusResp.Body.Close()

			// Read status response body
			statusBytes, err := io.ReadAll(statusResp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read status response: %v", err)
			}

			var status QueueResponse
			if err := json.Unmarshal(statusBytes, &status); err != nil {
				return nil, fmt.Errorf("failed to parse status response: %v", err)
			}

			switch status.Status {
			case "IN_QUEUE":
				if status.QueuePosition > 0 {
					// Send queue position update to user
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						sendPMMethod.Call([]reflect.Value{
							reflect.ValueOf(ctx),
							reflect.ValueOf(userNick),
							reflect.ValueOf(fmt.Sprintf("Your request is in queue, position: %d", status.QueuePosition)),
						})
					}
				}
			case "IN_PROGRESS":
				// Send progress update to user if we have logs
				if len(status.Logs) > 0 {
					// Get the latest log entry
					latestLog := status.Logs[len(status.Logs)-1]

					// Use reflection to call SendPM on the bot
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						sendPMMethod.Call([]reflect.Value{
							reflect.ValueOf(ctx),
							reflect.ValueOf(userNick),
							reflect.ValueOf(fmt.Sprintf("Progress: %s", latestLog.Message)),
						})
					}
				}
			case "COMPLETED":
				return &status, nil
			}
		}
	}
}

// notifyQueuePosition sends a notification about the queue position
func (c *Client) notifyQueuePosition(ctx context.Context, queueResp QueueResponse, bot interface{}, userNick string) {
	if queueResp.QueuePosition >= 0 {
		// Use reflection to call SendPM on the bot
		botValue := reflect.ValueOf(bot)
		sendPMMethod := botValue.MethodByName("SendPM")
		if sendPMMethod.IsValid() {
			message := "Your request is at the front of the queue."
			if queueResp.QueuePosition > 0 {
				message = fmt.Sprintf("Your request is in queue. Position: %d", queueResp.QueuePosition)
			}
			sendPMMethod.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(userNick),
				reflect.ValueOf(message),
			})
		}
	}
}
