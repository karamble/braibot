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
	"strings"
)

// GenerateSpeech generates speech from text
func (c *Client) GenerateSpeech(ctx context.Context, text string, voiceID string, bot interface{}, userNick string) (*AudioResponse, error) {
	// Create request body
	reqBody := map[string]interface{}{
		"text":     text,
		"voice_id": voiceID,
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/minimax-tts/text-to-speech", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the initial queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}
	resp.Body.Close()

	// Notify about queue position
	c.notifyQueuePosition(ctx, queueResp, bot, userNick)

	// Poll for completion
	_, err = c.pollQueueStatus(ctx, queueResp, bot, userNick)
	if err != nil {
		return nil, fmt.Errorf("failed to poll queue status: %v", err)
	}

	// Get final response using the base request URL (without /status)
	responseURL := strings.TrimSuffix(queueResp.ResponseURL, "/status")
	req, err := http.NewRequestWithContext(ctx, "GET", responseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create final response request: %v", err)
	}
	req.Header.Set("Authorization", "Key "+c.apiKey)

	finalResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %v", err)
	}
	defer finalResp.Body.Close()

	// Read final response body
	finalBytes, err := io.ReadAll(finalResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read final response: %v", err)
	}

	var finalResponse AudioResponse
	if err := json.Unmarshal(finalBytes, &finalResponse); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v", err)
	}

	return &finalResponse, nil
}
