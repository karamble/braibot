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

	kit "github.com/vctt94/bisonbotkit"
)

// GenerateVideoFromImage generates a video from an input image using the specified model
func (c *Client) GenerateVideoFromImage(ctx context.Context, prompt, imageURL, modelName string, bot *kit.Bot, userID string, duration int, aspectRatio string, negativePrompt string, cfgScale float64) (*VideoResponse, error) {
	// Create request body based on API schema
	reqBody := map[string]interface{}{
		"prompt":          prompt,
		"image_url":       imageURL,
		"duration":        fmt.Sprintf("%d", duration),
		"aspect_ratio":    aspectRatio,
		"negative_prompt": negativePrompt,
		"cfg_scale":       cfgScale,
	}

	// Debug: Print request body
	if c.debug {
		fmt.Printf("Video request body: %+v\n", reqBody)
	}

	// Make request to queue
	resp, err := c.makeRequest(ctx, "POST", "/kling-video/v2/master/image-to-video", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}

	// Debug: Print parsed queue response
	if c.debug {
		fmt.Printf("Parsed queue response: %+v\n", queueResp)
	}

	// Notify user of queue position
	c.notifyQueuePosition(ctx, queueResp, bot, userID)

	// Poll for completion
	_, err = c.pollQueueStatus(ctx, queueResp, bot, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for completion: %v", err)
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

	// Debug: Print the raw response
	if c.debug {
		fmt.Printf("Raw video response: %s\n", string(finalBytes))
	}

	var finalResponse VideoResponse
	if err := json.Unmarshal(finalBytes, &finalResponse); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v", err)
	}

	// Debug: Print parsed final response
	if c.debug {
		fmt.Printf("Parsed video response: %+v\n", finalResponse)
	}

	return &finalResponse, nil
}
