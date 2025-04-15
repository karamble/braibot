// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package falapi

import "encoding/json"

// Model represents a Fal.ai model configuration
type Model struct {
	Name        string  // Name of the model
	Description string  // Description of the model
	Price       float64 // Price per picture in USD
}

// QueueResponse represents the response from the queue API
type QueueResponse struct {
	Status      string `json:"status"`
	RequestID   string `json:"request_id"`
	ResponseURL string `json:"response_url"`
	StatusURL   string `json:"status_url"`
	CancelURL   string `json:"cancel_url"`
	Logs        []struct {
		Message   string `json:"message"`
		Level     string `json:"level"`
		Source    string `json:"source"`
		Timestamp string `json:"timestamp"`
	} `json:"logs"`
	QueuePosition int `json:"queue_position"`
}

// ImageResponse represents the final image generation response
type ImageResponse struct {
	Images []struct {
		URL         string `json:"url"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		ContentType string `json:"content_type"`
	} `json:"images"`
	Timings struct {
		Inference float64 `json:"inference"`
	} `json:"timings"`
	Seed            json.Number `json:"seed"`
	HasNSFWConcepts []bool      `json:"has_nsfw_concepts"`
	Prompt          string      `json:"prompt"`
}

// AudioResponse represents the final audio generation response
type AudioResponse struct {
	Audio struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		FileName    string `json:"file_name"`
		FileSize    int    `json:"file_size"`
	} `json:"audio"`
	DurationMs int `json:"duration_ms"`
}

// StatusResponse represents the status check response
type StatusResponse struct {
	Status        string `json:"status"`
	QueuePosition int    `json:"queue_position,omitempty"`
}
