// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// pollQueueStatus polls the queue status until completion or error
func (c *Client) pollQueueStatus(ctx context.Context, queueResp QueueResponse, progress ProgressCallback) (*QueueResponse, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastPosition := queueResp.Position
	lastETA := queueResp.ETA

	// Construct the status URL with logs parameter
	statusURL := queueResp.ResponseURL + "/status?logs=1"
	if c.debug {
		fmt.Printf("DEBUG - Initial status URL: %s\n", statusURL)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Create request to check status
			req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create status request: %v", err)
			}
			req.Header.Set("Authorization", "Key "+c.apiKey)

			// Make request
			resp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to check status: %v", err)
			}

			if c.debug {
				fmt.Printf("DEBUG - Queue Status Poll:\n")
				fmt.Printf("  URL: %s\n", statusURL)
				fmt.Printf("  Status Code: %d\n", resp.StatusCode)
			}

			// Read the response body
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %v", err)
			}

			if c.debug {
				fmt.Printf("  Response Body: %s\n", string(body))
			}

			// Check for HTTP errors (excluding 202 Accepted)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
				return nil, fmt.Errorf("queue status check failed with status code: %d, response: %s", resp.StatusCode, string(body))
			}

			// Parse response
			var statusResp struct {
				QueueResponse
				Logs []struct {
					Message   string `json:"message"`
					Level     string `json:"level"`
					Source    string `json:"source"`
					Timestamp string `json:"timestamp"`
				} `json:"logs"`
			}
			if err := json.Unmarshal(body, &statusResp); err != nil {
				return nil, fmt.Errorf("failed to decode status response: %v", err)
			}

			if c.debug {
				fmt.Printf("  Queue ID: %s\n", statusResp.QueueID)
				fmt.Printf("  Status: %s\n", statusResp.Status)
				fmt.Printf("  Position: %d\n", statusResp.Position)
				fmt.Printf("  ETA: %d seconds\n", statusResp.ETA)
				if len(statusResp.Logs) > 0 {
					fmt.Printf("  Logs:\n")
					for _, log := range statusResp.Logs {
						fmt.Printf("    [%s] %s: %s\n", log.Timestamp, log.Level, log.Message)
					}
				}
			}

			// Send log messages to the progress callback
			if progress != nil && len(statusResp.Logs) > 0 {
				for _, log := range statusResp.Logs {
					progress.OnLogMessage(log.Message)
				}
			}

			// Check for completion
			if statusResp.Status == "COMPLETED" {
				if c.debug {
					fmt.Printf("DEBUG - Queue completed successfully\n")
				}
				// Set the base URL for fetching the final result
				statusResp.ResponseURL = strings.TrimSuffix(statusURL, "/status?logs=1")
				return &statusResp.QueueResponse, nil
			}

			// Check for error
			if statusResp.Status == "FAILED" {
				if c.debug {
					fmt.Printf("DEBUG - Queue failed\n")
				}
				return nil, &Error{
					Code:    "GENERATION_FAILED",
					Message: "image generation failed",
				}
			}

			// Notify about status changes
			if progress != nil {
				progress.OnProgress(statusResp.Status)
			}

			// Notify progress if position or ETA changed
			if progress != nil && (statusResp.Position != lastPosition || statusResp.ETA != lastETA) {
				if c.debug {
					fmt.Printf("DEBUG - Queue progress update:\n")
					fmt.Printf("  Position changed: %d -> %d\n", lastPosition, statusResp.Position)
					fmt.Printf("  ETA changed: %d -> %d seconds\n", lastETA, statusResp.ETA)
				}
				progress.OnQueueUpdate(statusResp.Position, time.Duration(statusResp.ETA)*time.Second)
				lastPosition = statusResp.Position
				lastETA = statusResp.ETA
			}
		}
	}
}

// notifyQueuePosition sends a queue position update through the progress callback
func (c *Client) notifyQueuePosition(ctx context.Context, queueResp QueueResponse, progress ProgressCallback) {
	if progress != nil {
		progress.OnQueueUpdate(queueResp.Position, time.Duration(queueResp.ETA)*time.Second)
	}
}
