// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// GenerateImage generates an image from a text prompt
func (c *Client) GenerateImage(ctx context.Context, prompt string, modelName string, bot interface{}, userNick string) (*ImageResponse, error) {
	// Create request body
	reqBody := map[string]interface{}{
		"prompt": prompt,
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/"+modelName, reqBody)
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

	var finalResponse ImageResponse
	if err := json.Unmarshal(finalBytes, &finalResponse); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v", err)
	}

	if len(finalResponse.Images) == 0 {
		return nil, fmt.Errorf("no images in response")
	}

	// Download the image from the URL
	imgResp, err := http.Get(finalResponse.Images[0].URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer imgResp.Body.Close()

	// Read the image data
	imgData, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v", err)
	}

	// Encode the image as base64
	base64Image := base64.StdEncoding.EncodeToString(imgData)

	// Create the PM message format
	_ = fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--",
		url.QueryEscape(prompt),
		finalResponse.Images[0].ContentType,
		base64Image)

	return &finalResponse, nil
}
