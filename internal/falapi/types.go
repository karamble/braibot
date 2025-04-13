package falapi

import "encoding/json"

// Model represents a Fal.ai model configuration
type Model struct {
	Name        string  // Name of the model
	Description string  // Description of the model
	Price       float64 // Price per picture in USD
}

// FalResponse represents the response from Fal.ai API
type FalResponse struct {
	Status        string `json:"status,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	ResponseURL   string `json:"response_url,omitempty"`
	StatusURL     string `json:"status_url,omitempty"`
	CancelURL     string `json:"cancel_url,omitempty"`
	QueuePosition int    `json:"queue_position,omitempty"`
	Logs          []struct {
		Message   string `json:"message"`
		Level     string `json:"level"`
		Source    string `json:"source"`
		Timestamp string `json:"timestamp"`
	} `json:"logs,omitempty"`
	Response struct {
		Images []struct {
			URL         string `json:"url"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			ContentType string `json:"content_type"`
		} `json:"images"`
	} `json:"response,omitempty"`
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
