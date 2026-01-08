// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"context"
	"encoding/json"
	"fmt"
)

// Transcribe transcribes audio to text using Scribe V2
func (c *Client) Transcribe(ctx context.Context, req *ScribeV2Request) (*ScribeV2Response, error) {
	const modelName = "elevenlabs/speech-to-text/scribe-v2"
	const endpoint = "/elevenlabs/speech-to-text/scribe-v2"

	// Validate audio_url is provided
	if req.AudioURL == "" {
		return nil, fmt.Errorf("audio_url is required")
	}

	// Validate options
	currentOpts := ScribeV2Options{
		Task:        req.Task,
		Language:    req.Language,
		ChunkLevel:  req.ChunkLevel,
		Diarize:     req.Diarize,
		NumSpeakers: req.NumSpeakers,
	}
	if err := currentOpts.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
	}

	// Extract progress callback
	var progress ProgressCallback
	if req.Progress != nil {
		progress = req.Progress
	}

	// Build request body
	reqBody := map[string]interface{}{
		"audio_url": req.AudioURL,
	}

	// Add optional parameters
	if req.Task != "" {
		reqBody["task"] = req.Task
	}
	if req.Language != "" {
		reqBody["language"] = req.Language
	}
	if req.ChunkLevel != "" {
		reqBody["chunk_level"] = req.ChunkLevel
	}
	if req.Diarize != nil {
		reqBody["diarize"] = *req.Diarize
	}
	if req.NumSpeakers != nil {
		reqBody["num_speakers"] = *req.NumSpeakers
	}

	// Validate the requested model
	if _, exists := GetModel(modelName, "audio2text"); !exists {
		return nil, &Error{
			Code:    "INVALID_MODEL",
			Message: fmt.Sprintf("invalid or unsupported model %s for audio2text", modelName),
		}
	}

	// Define the decoder for the transcription response
	decodeFunc := func(data []byte) (interface{}, error) {
		var response ScribeV2Response

		if err := json.Unmarshal(data, &response); err != nil {
			return nil, fmt.Errorf("failed to parse transcription response: %w. Body: %s", err, string(data))
		}

		return &response, nil
	}

	// Execute the workflow
	result, err := c.executeAsyncWorkflow(ctx, endpoint, reqBody, progress, decodeFunc)
	if err != nil {
		return nil, err
	}

	return result.(*ScribeV2Response), nil
}
