// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
)

// GenerateSpeech generates speech from text using the specified model
func (c *Client) GenerateSpeech(ctx context.Context, req SpeechRequest) (*AudioResponse, error) {
	// Validate the requested model
	if _, exists := GetModel(req.Model, "text2speech"); !exists {
		return nil, &Error{
			Code:    "INVALID_MODEL",
			Message: "invalid or unsupported model for text2speech: " + req.Model,
		}
	}

	// Use the model name to form the endpoint path
	// Assumes the endpoint path directly corresponds to the model name for now.
	// More complex routing might be needed if this assumption changes.
	endpoint := "/" + req.Model

	// Base request body
	reqBody := map[string]interface{}{
		"text": req.Text,
	}

	// Add model-specific parameters. Currently handling voice_id for minimax.
	if req.Model == "minimax-tts/text-to-speech" && req.VoiceID != "" {
		reqBody["voice_id"] = req.VoiceID
	}

	// Add any additional options from the request
	for k, v := range req.Options {
		// Avoid overwriting fields already set (text, voice_id for minimax)
		if _, exists := reqBody[k]; !exists {
			reqBody[k] = v
		}
	}

	// Define the decoder for the final audio response
	decodeFunc := func(data []byte) (interface{}, error) {
		var response struct {
			Audio struct {
				URL         string `json:"url"`
				ContentType string `json:"content_type"`
				FileName    string `json:"file_name"`
				FileSize    int    `json:"file_size"`
			} `json:"audio"`
			Duration float64 `json:"duration"` // Fal uses duration now
		}

		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to parse final audio response: %w. Body: %s", err, string(data))
		}

		if response.Audio.URL == "" {
			return nil, &Error{
				Code:    "NO_AUDIO_URL",
				Message: "no audio URL found in response",
			}
		}

		contentType := response.Audio.ContentType
		if contentType == "" {
			contentType = "audio/mpeg" // Default if missing
		}

		return &AudioResponse{
			AudioURL:    response.Audio.URL,
			ContentType: contentType,
			FileName:    response.Audio.FileName,
			FileSize:    response.Audio.FileSize,
			Duration:    response.Duration, // Use the float duration
		}, nil
	}

	// Execute the workflow
	result, err := c.executeAsyncWorkflow(ctx, endpoint, reqBody, req.Progress, decodeFunc)
	if err != nil {
		return nil, err // Error already wrapped
	}

	return result.(*AudioResponse), nil
}
