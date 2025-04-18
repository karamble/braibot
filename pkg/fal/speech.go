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
// It accepts specific request types like *MinimaxTTSRequest.
func (c *Client) GenerateSpeech(ctx context.Context, req interface{}) (*AudioResponse, error) {
	var modelName string
	var endpoint string
	var reqBody map[string]interface{}
	var progress ProgressCallback

	// Extract progress callback if available
	if progressable, ok := req.(Progressable); ok {
		progress = progressable.GetProgress()
	}

	// Determine model name, endpoint and create request body based on request type
	switch r := req.(type) {
	case *MinimaxTTSRequest:
		modelName = "minimax-tts/text-to-speech"
		endpoint = "/" + modelName // Assuming endpoint matches model name

		// Validate options based on request values
		currentOpts := MinimaxTTSOptions{
			Speed:      r.Speed,
			Vol:        r.Vol,
			Pitch:      r.Pitch,
			Emotion:    r.Emotion,
			SampleRate: r.SampleRate,
			Bitrate:    r.Bitrate,
			Format:     r.Format,
			Channel:    r.Channel,
		}
		if err := currentOpts.Validate(); err != nil {
			return nil, fmt.Errorf("invalid options for %s: %v", modelName, err)
		}

		// Build nested request body structure
		voiceSetting := make(map[string]interface{})
		voiceSetting["voice_id"] = r.VoiceID // Always include voice ID
		if r.Speed != nil {
			voiceSetting["speed"] = *r.Speed
		}
		if r.Vol != nil {
			voiceSetting["vol"] = *r.Vol
		}
		if r.Pitch != nil {
			voiceSetting["pitch"] = *r.Pitch
		}
		if r.Emotion != "" {
			voiceSetting["emotion"] = r.Emotion
		}

		audioSetting := make(map[string]interface{})
		if r.SampleRate != "" {
			audioSetting["sample_rate"] = r.SampleRate
		}
		if r.Bitrate != "" {
			audioSetting["bitrate"] = r.Bitrate
		}
		if r.Format != "" {
			audioSetting["format"] = r.Format
		}
		if r.Channel != "" {
			audioSetting["channel"] = r.Channel
		}

		reqBody = map[string]interface{}{"text": r.Text}
		if len(voiceSetting) > 1 { // Only add if more than just voice_id
			reqBody["voice_setting"] = voiceSetting
		}
		if len(audioSetting) > 0 {
			reqBody["audio_setting"] = audioSetting
		}

		r.Model = modelName // Set model name internally
	// Add cases for other specific speech requests here
	// case *OtherSpeechRequest:
	// ...
	default:
		return nil, fmt.Errorf("unsupported speech request type: %T", req)
	}

	// Validate the requested model
	if _, exists := GetModel(modelName, "text2speech"); !exists {
		return nil, &Error{
			Code:    "INVALID_MODEL",
			Message: fmt.Sprintf("invalid or unsupported model %s for text2speech", modelName),
		}
	}

	// Add any additional generic options from the request
	if optionsGetter, ok := req.(interface{ GetOptions() map[string]interface{} }); ok {
		for k, v := range optionsGetter.GetOptions() {
			// Avoid overwriting fields already set by specific request types
			if _, exists := reqBody[k]; !exists {
				reqBody[k] = v
			}
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
	result, err := c.executeAsyncWorkflow(ctx, endpoint, reqBody, progress, decodeFunc)
	if err != nil {
		return nil, err // Error already wrapped
	}

	return result.(*AudioResponse), nil
}
