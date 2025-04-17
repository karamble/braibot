// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"fmt"
	"strconv"
	"time"
)

// ProgressCallback is an interface for receiving progress updates
type ProgressCallback interface {
	OnQueueUpdate(position int, eta time.Duration)
	OnLogMessage(message string)
	OnProgress(status string)
	OnError(err error)
}

// ModelOptions represents the common interface for all model options
type ModelOptions interface {
	GetDefaultValues() map[string]interface{}
	Validate() error
}

// Veo2Options represents the options available for the Veo2 model
type Veo2Options struct {
	AspectRatio string `json:"aspect_ratio,omitempty"` // 16:9, 9:16, 1:1
	Duration    string `json:"duration,omitempty"`     // 5, 6, 7, 8
}

// GetDefaultValues returns the default values for Veo2 options
func (o *Veo2Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"aspect_ratio": "16:9",
		"duration":     "5",
	}
}

// Validate validates the Veo2 options
func (o *Veo2Options) Validate() error {
	validAspectRatios := map[string]bool{
		"16:9": true,
		"9:16": true,
		"1:1":  true,
	}
	validDurations := map[string]bool{
		"5": true,
		"6": true,
		"7": true,
		"8": true,
	}

	if o.AspectRatio != "" && !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect ratio: %s (must be one of: 16:9, 9:16, 1:1)", o.AspectRatio)
	}
	if o.Duration != "" && !validDurations[o.Duration] {
		return fmt.Errorf("invalid duration: %s (must be one of: 5, 6, 7, 8)", o.Duration)
	}
	return nil
}

// KlingVideoOptions represents the options available for the Kling-video model
type KlingVideoOptions struct {
	Duration       string  `json:"duration,omitempty"`     // Duration in seconds
	AspectRatio    string  `json:"aspect_ratio,omitempty"` // 16:9, 9:16
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	CFGScale       float64 `json:"cfg_scale,omitempty"`
}

// GetDefaultValues returns the default values for Kling-video options
func (o *KlingVideoOptions) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"duration":        "5",
		"aspect_ratio":    "16:9",
		"negative_prompt": "blur, distort, and low quality",
		"cfg_scale":       0.5,
	}
}

// Validate validates the Kling-video options
func (o *KlingVideoOptions) Validate() error {
	validAspectRatios := map[string]bool{
		"16:9": true,
		"9:16": true,
	}

	if o.AspectRatio != "" && !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect ratio: %s", o.AspectRatio)
	}
	if o.Duration != "" {
		dur, err := strconv.Atoi(o.Duration)
		if err != nil || dur < 5 {
			return fmt.Errorf("invalid duration: %s (must be at least 5 seconds)", o.Duration)
		}
	}
	if o.CFGScale < 0 || o.CFGScale > 1 {
		return fmt.Errorf("invalid cfg_scale: %f (must be between 0 and 1)", o.CFGScale)
	}
	return nil
}

// Model represents a Fal.ai model
type Model struct {
	Name        string
	Description string
	PriceUSD    float64
	Type        string
	HelpDoc     string
	Options     ModelOptions // Model-specific options
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

// Progressable is an interface for types that can provide progress updates
type Progressable interface {
	GetProgress() ProgressCallback
}

// BaseVideoRequest contains common fields for all video generation requests
type BaseVideoRequest struct {
	Prompt   string                 `json:"prompt"`
	ImageURL string                 `json:"image_url"`
	Model    string                 `json:"-"`
	Options  map[string]interface{} `json:"-"`
	Progress ProgressCallback
}

// GetProgress returns the progress callback
func (r *BaseVideoRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// GetOptions returns the options map
func (r *BaseVideoRequest) GetOptions() map[string]interface{} {
	return r.Options
}

// Veo2Request represents a request to generate a video using the Veo2 model
type Veo2Request struct {
	BaseVideoRequest
	Duration    string `json:"duration,omitempty"`
	AspectRatio string `json:"aspect_ratio,omitempty"`
}

// KlingVideoRequest represents a request to generate a video using the Kling-video model
type KlingVideoRequest struct {
	BaseVideoRequest
	Duration       string  `json:"duration,omitempty"`
	AspectRatio    string  `json:"aspect_ratio,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	CFGScale       float64 `json:"cfg_scale,omitempty"`
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

// StarVectorRequest represents a request to generate an SVG using the star-vector model
type StarVectorRequest struct {
	ImageURL string                 `json:"image_url"`
	Options  map[string]interface{} `json:"-"`
	Progress ProgressCallback
}

// GetProgress returns the progress callback
func (r *StarVectorRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// GetOptions returns the options map
func (r *StarVectorRequest) GetOptions() map[string]interface{} {
	return r.Options
}

// StarVectorResponse represents the response from the star-vector model
type StarVectorResponse struct {
	SVG struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		FileName    string `json:"file_name"`
		FileSize    int    `json:"file_size"`
	} `json:"svg"`
}
