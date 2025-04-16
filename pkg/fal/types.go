// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"time"
)

// ProgressCallback is an interface for receiving progress updates
type ProgressCallback interface {
	OnQueueUpdate(position int, eta time.Duration)
	OnLogMessage(message string)
	OnProgress(status string)
	OnError(err error)
}

// Model represents a Fal.ai model
type Model struct {
	Name        string
	Description string
	PriceUSD    float64
	Type        string
	HelpDoc     string
}

// ImageRequest represents a request to generate an image
type ImageRequest struct {
	Prompt   string
	Model    string
	Options  map[string]interface{}
	Progress ProgressCallback
}

// ImageResponse represents the response from an image generation request
type ImageResponse struct {
	Images []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	} `json:"images"`
	NSFW        bool      `json:"nsfw"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// SpeechRequest represents a request to generate speech
type SpeechRequest struct {
	Text     string
	VoiceID  string
	Options  map[string]interface{}
	Progress ProgressCallback
}

// AudioResponse represents the response from a speech generation request
type AudioResponse struct {
	AudioURL    string  `json:"audio_url"`
	ContentType string  `json:"content_type"`
	FileName    string  `json:"file_name"`
	FileSize    int     `json:"file_size"`
	Duration    float64 `json:"duration"`
}

// VideoRequest represents a request to generate a video using the kling-video model
type VideoRequest struct {
	Prompt         string                 `json:"prompt"`
	ImageURL       string                 `json:"image_url"`
	Duration       string                 `json:"duration,omitempty"`
	AspectRatio    string                 `json:"aspect_ratio,omitempty"`
	NegativePrompt string                 `json:"negative_prompt,omitempty"`
	CFGScale       float64                `json:"cfg_scale,omitempty"`
	Options        map[string]interface{} `json:"-"`
	Progress       ProgressCallback
}

// VideoResponse represents the response from the kling-video model
type VideoResponse struct {
	// Format 1: {"video": {"url": "..."}}
	Video struct {
		URL string `json:"url"`
	} `json:"video"`

	// Format 2: {"url": "..."}
	URL string `json:"url"`

	// Format 3: {"video_url": "..."}
	VideoURL string `json:"video_url"`
}

// GetURL returns the video URL from any of the possible fields
func (r *VideoResponse) GetURL() string {
	if r.Video.URL != "" {
		return r.Video.URL
	}
	if r.URL != "" {
		return r.URL
	}
	return r.VideoURL
}

// QueueResponse represents the response from a queue request
type QueueResponse struct {
	ResponseURL string `json:"response_url"`
	QueueID     string `json:"queue_id"`
	Status      string `json:"status"`
	Position    int    `json:"position"`
	ETA         int    `json:"eta"`
}

// Error represents a Fal.ai API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
