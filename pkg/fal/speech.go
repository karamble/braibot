// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// GenerateSpeech generates speech from text
func (c *Client) GenerateSpeech(ctx context.Context, req SpeechRequest) (*AudioResponse, error) {
	// Validate model
	if _, exists := GetModel(req.VoiceID, "text2speech"); !exists {
		return nil, &Error{
			Code:    "INVALID_MODEL",
			Message: "invalid model for text2speech: " + req.VoiceID,
		}
	}

	// Create request body
	reqBody := map[string]interface{}{
		"text": req.Text,
	}

	// Add any additional options
	for k, v := range req.Options {
		reqBody[k] = v
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/"+req.VoiceID, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the initial queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}

	// Notify about queue position
	c.notifyQueuePosition(ctx, queueResp, req.Progress)

	// Poll for completion
	_, err = c.pollQueueStatus(ctx, queueResp, req.Progress)
	if err != nil {
		return nil, fmt.Errorf("failed to poll queue status: %v", err)
	}

	// Get final response
	finalResp, err := c.makeRequest(ctx, "GET", queueResp.ResponseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get final response: %v", err)
	}
	defer finalResp.Body.Close()

	// Read final response body
	finalBytes, err := io.ReadAll(finalResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read final response: %v", err)
	}

	// Debug log the response
	if c.debug {
		fmt.Printf("DEBUG - Final response body: %s\n", string(finalBytes))
	}

	// Parse the response
	var response struct {
		Audio struct {
			URL         string `json:"url"`
			ContentType string `json:"content_type"`
			FileName    string `json:"file_name"`
			FileSize    int    `json:"file_size"`
		} `json:"audio"`
		DurationMs int `json:"duration_ms"`
	}

	if err := json.Unmarshal(finalBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse final response: %v", err)
	}

	if response.Audio.URL == "" {
		return nil, &Error{
			Code:    "NO_AUDIO",
			Message: "no audio URL in response",
		}
	}

	// Set default content type if not provided
	contentType := response.Audio.ContentType
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	return &AudioResponse{
		AudioURL:    response.Audio.URL,
		ContentType: contentType,
	}, nil
}
