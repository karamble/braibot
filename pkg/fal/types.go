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
// This interface is used for compile-time type safety and generic handling.
type ModelOptions interface {
	GetDefaultValues() map[string]interface{}
	Validate() error
}

// Veo2Options represents the options available for the Veo2 model
type Veo2Options struct {
	AspectRatio string `json:"aspect_ratio,omitempty"` // auto, auto_prefer_portrait, 16:9, 9:16
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
		"auto":                 true,
		"auto_prefer_portrait": true,
		"16:9":                 true,
		"9:16":                 true,
	}
	validDurations := map[string]bool{
		"5": true,
		"6": true,
		"7": true,
		"8": true,
	}

	if o.AspectRatio != "" && !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect ratio: %s (must be one of: auto, auto_prefer_portrait, 16:9, 9:16)", o.AspectRatio)
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

// MinimaxDirectorOptions represents the options available for the minimax-video-01-director model
type MinimaxDirectorOptions struct {
	PromptOptimizer *bool `json:"prompt_optimizer,omitempty"` // Default: true
}

// GetDefaultValues returns the default values for MinimaxDirector options
func (o *MinimaxDirectorOptions) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"prompt_optimizer": &defaultOptimizer,
	}
}

// Validate validates the MinimaxDirector options
func (o *MinimaxDirectorOptions) Validate() error {
	// No specific validation needed for a boolean flag yet
	return nil
}

// FluxSchnellOptions represents the options available for the fal-ai/flux/schnell model
type FluxSchnellOptions struct {
	ImageSize           string  `json:"image_size,omitempty"`            // square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`   // Default: 4
	Seed                *int    `json:"seed,omitempty"`                  // Optional seed
	SyncMode            bool    `json:"sync_mode,omitempty"`             // Default: false
	NumImages           int     `json:"num_images,omitempty"`            // Default: 1
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"` // Default: true
	CFGScale            float64 `json:"cfg_scale,omitempty"`
	PromptOptimizer     *bool   `json:"prompt_optimizer,omitempty"` // Mirrored option
}

// GetDefaultValues returns the default values for Flux Schnell options
func (o *FluxSchnellOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	return map[string]interface{}{
		"image_size":            "landscape_4_3",
		"num_inference_steps":   4,
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker, // Use pointer for default true bool
		// seed and sync_mode default to nil/false implicitly
	}
}

// Validate validates the Flux Schnell options
func (o *FluxSchnellOptions) Validate() error {
	validImageSizes := map[string]bool{
		"square_hd":      true,
		"square":         true,
		"portrait_4_3":   true,
		"portrait_16_9":  true,
		"landscape_4_3":  true,
		"landscape_16_9": true,
	}

	if o.ImageSize != "" && !validImageSizes[o.ImageSize] {
		// TODO: Add support for {width, height} object validation if needed
		return fmt.Errorf("invalid image_size: %s (must be one of: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9)", o.ImageSize)
	}
	if o.NumInferenceSteps < 0 {
		return fmt.Errorf("num_inference_steps cannot be negative: %d", o.NumInferenceSteps)
	}
	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}

	return nil
}

// FluxProV1_1Options represents the options available for the fal-ai/flux-pro/v1.1 model
type FluxProV1_1Options struct {
	ImageSize           string `json:"image_size,omitempty"`            // square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9
	Seed                *int   `json:"seed,omitempty"`                  // Optional seed
	SyncMode            bool   `json:"sync_mode,omitempty"`             // Default: false
	NumImages           int    `json:"num_images,omitempty"`            // Default: 1
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"` // Default: true
	SafetyTolerance     string `json:"safety_tolerance,omitempty"`      // Enum: 1, 2, 3, 4, 5, 6. Default: "2"
	OutputFormat        string `json:"output_format,omitempty"`         // Enum: jpeg, png. Default: "jpeg"
}

// GetDefaultValues returns the default values for Flux Pro v1.1 options
func (o *FluxProV1_1Options) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	return map[string]interface{}{
		"image_size":            "landscape_4_3",
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker,
		"safety_tolerance":      "2",
		"output_format":         "jpeg",
		// seed and sync_mode default to nil/false implicitly
	}
}

// Validate validates the Flux Pro v1.1 options
func (o *FluxProV1_1Options) Validate() error {
	validImageSizes := map[string]bool{
		"square_hd":      true,
		"square":         true,
		"portrait_4_3":   true,
		"portrait_16_9":  true,
		"landscape_4_3":  true,
		"landscape_16_9": true,
	}
	validSafetyTolerances := map[string]bool{
		"1": true, "2": true, "3": true, "4": true, "5": true, "6": true,
	}
	validOutputFormats := map[string]bool{
		"jpeg": true, "png": true,
	}

	if o.ImageSize != "" && !validImageSizes[o.ImageSize] {
		return fmt.Errorf("invalid image_size: %s", o.ImageSize)
	}
	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}
	if o.SafetyTolerance != "" && !validSafetyTolerances[o.SafetyTolerance] {
		return fmt.Errorf("invalid safety_tolerance: %s (must be 1-6)", o.SafetyTolerance)
	}
	if o.OutputFormat != "" && !validOutputFormats[o.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s (must be jpeg or png)", o.OutputFormat)
	}
	return nil
}

// MinimaxTTSOptions represents options for the minimax-tts model.
type MinimaxTTSOptions struct {
	// Voice Settings
	// VoiceID handled in request struct
	Speed   *float64 `json:"speed,omitempty"`   // 0.5 - 2.0, default 1.0
	Vol     *float64 `json:"vol,omitempty"`     // 0 - 10, default 1.0
	Pitch   *int     `json:"pitch,omitempty"`   // -12 - 12, optional
	Emotion string   `json:"emotion,omitempty"` // happy, sad, etc. optional
	// Audio Settings
	SampleRate string `json:"sample_rate,omitempty"` // 8000, 16000, ..., default 32000
	Bitrate    string `json:"bitrate,omitempty"`     // 32000, 64000, ..., default 128000
	Format     string `json:"format,omitempty"`      // mp3, pcm, flac, default mp3
	Channel    string `json:"channel,omitempty"`     // 1 (mono), 2 (stereo), default 1
}

// GetDefaultValues returns default values for MinimaxTTSOptions.
func (o *MinimaxTTSOptions) GetDefaultValues() map[string]interface{} {
	// Define defaults as pointers where applicable to match struct fields
	defaultSpeed := 1.0
	defaultVol := 1.0
	return map[string]interface{}{
		"speed":       &defaultSpeed,
		"vol":         &defaultVol,
		"sample_rate": "32000",
		"bitrate":     "128000",
		"format":      "mp3",
		"channel":     "1",
		// Pitch and Emotion default to nil/""
	}
}

// Validate validates MinimaxTTSOptions.
func (o *MinimaxTTSOptions) Validate() error {
	// Voice Settings Validation
	if o.Speed != nil && (*o.Speed < 0.5 || *o.Speed > 2.0) {
		return fmt.Errorf("invalid speed: %f (must be between 0.5 and 2.0)", *o.Speed)
	}
	if o.Vol != nil && (*o.Vol < 0 || *o.Vol > 10.0) {
		return fmt.Errorf("invalid vol: %f (must be between 0 and 10)", *o.Vol)
	}
	if o.Pitch != nil && (*o.Pitch < -12 || *o.Pitch > 12) {
		return fmt.Errorf("invalid pitch: %d (must be between -12 and 12)", *o.Pitch)
	}
	validEmotions := map[string]bool{"happy": true, "sad": true, "angry": true, "fearful": true, "disgusted": true, "surprised": true, "neutral": true, "": true}
	if !validEmotions[o.Emotion] {
		return fmt.Errorf("invalid emotion: %s", o.Emotion)
	}
	// Audio Settings Validation
	validSampleRates := map[string]bool{"8000": true, "16000": true, "22050": true, "24000": true, "32000": true, "44100": true, "": true}
	if !validSampleRates[o.SampleRate] {
		return fmt.Errorf("invalid sample_rate: %s", o.SampleRate)
	}
	validBitrates := map[string]bool{"32000": true, "64000": true, "128000": true, "256000": true, "": true}
	if !validBitrates[o.Bitrate] {
		return fmt.Errorf("invalid bitrate: %s", o.Bitrate)
	}
	validFormats := map[string]bool{"mp3": true, "pcm": true, "flac": true, "": true}
	if !validFormats[o.Format] {
		return fmt.Errorf("invalid format: %s", o.Format)
	}
	validChannels := map[string]bool{"1": true, "2": true, "": true}
	if !validChannels[o.Channel] {
		return fmt.Errorf("invalid channel: %s", o.Channel)
	}
	return nil
}

// ModelDefinition is an interface for types that define a Fal.ai model.
// Implementations of this interface are expected to register themselves
// using the registerModel function in their init() function.
type ModelDefinition interface {
	Define() Model
}

// Model represents a Fal.ai model
type Model struct {
	Name             string
	Description      string
	PriceUSD         float64
	Type             string
	HelpDoc          string
	Options          interface{} // Model-specific options
	PerSecondPricing bool        // Indicates if model uses per-second billing
}

// BaseImageRequest represents the base fields for an image generation request
// (text2image or image2image)
type BaseImageRequest struct {
	Prompt   string                 `json:"prompt,omitempty"`    // Optional for image2image
	ImageURL string                 `json:"image_url,omitempty"` // Required for image2image
	Model    string                 `json:"-"`                   // Internal use: model name
	Options  map[string]interface{} `json:"-"`                   // Fallback for generic options
	Progress ProgressCallback       `json:"-"`                   // Progress callback interface
}

// GetProgress returns the progress callback
func (r *BaseImageRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// GetOptions returns the options map
func (r *BaseImageRequest) GetOptions() map[string]interface{} {
	return r.Options
}

// FastSDXLRequest represents a request to generate an image using fast-sdxl
type FastSDXLRequest struct {
	BaseImageRequest
	NumImages int `json:"num_images,omitempty"` // Optional: Number of images to generate
}

// GhiblifyRequest represents a request to transform an image using ghiblify
type GhiblifyRequest struct {
	BaseImageRequest // Requires ImageURL to be set
	// No specific options for Ghiblify identified yet
}

// FluxSchnellRequest represents a request to generate an image using fal-ai/flux/schnell
type FluxSchnellRequest struct {
	BaseImageRequest
	ImageSize           string `json:"image_size,omitempty"`
	NumInferenceSteps   int    `json:"num_inference_steps,omitempty"`
	Seed                *int   `json:"seed,omitempty"`
	SyncMode            bool   `json:"sync_mode,omitempty"`
	NumImages           int    `json:"num_images,omitempty"`
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`
}

// FluxProV1_1Request represents a request for the fal-ai/flux-pro/v1.1 model
type FluxProV1_1Request struct {
	BaseImageRequest
	ImageSize           string `json:"image_size,omitempty"`
	Seed                *int   `json:"seed,omitempty"`
	SyncMode            bool   `json:"sync_mode,omitempty"`
	NumImages           int    `json:"num_images,omitempty"`
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`
	SafetyTolerance     string `json:"safety_tolerance,omitempty"`
	OutputFormat        string `json:"output_format,omitempty"`
}

// ImageOutput represents a single image result within an ImageResponse
type ImageOutput struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// ImageResponse represents the response from an image generation request
type ImageResponse struct {
	Images      []ImageOutput `json:"images"`
	NSFW        bool          `json:"nsfw"`
	CreatedAt   time.Time     `json:"created_at"`
	CompletedAt time.Time     `json:"completed_at"`
	Seed        uint64        `json:"seed"`
}

// BaseSpeechRequest represents the base fields for a speech generation request
type BaseSpeechRequest struct {
	Model    string                 `json:"-"` // Internal use: model name
	Text     string                 `json:"text"`
	Options  map[string]interface{} `json:"-"` // Fallback for generic options
	Progress ProgressCallback       `json:"-"` // Progress callback interface
}

// GetProgress returns the progress callback
func (r *BaseSpeechRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// GetOptions returns the options map
func (r *BaseSpeechRequest) GetOptions() map[string]interface{} {
	return r.Options
}

// MinimaxTTSRequest represents a request to generate speech using minimax-tts
type MinimaxTTSRequest struct {
	BaseSpeechRequest
	VoiceID string `json:"-"` // Not sent directly, part of voice_setting
	// Mirrored Options for convenience
	Speed      *float64 `json:"-"`
	Vol        *float64 `json:"-"`
	Pitch      *int     `json:"-"`
	Emotion    string   `json:"-"`
	SampleRate string   `json:"-"`
	Bitrate    string   `json:"-"`
	Format     string   `json:"-"`
	Channel    string   `json:"-"`
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

// QueueInfoable is an interface for types that can receive queue info callbacks
type QueueInfoable interface {
	GetQueueInfo() QueueInfoCallback
}

// BaseVideoRequest contains common fields for all video generation requests
type BaseVideoRequest struct {
	Prompt    string                 `json:"prompt"`
	ImageURL  string                 `json:"image_url"`
	Model     string                 `json:"-"`
	Options   map[string]interface{} `json:"-"`
	Progress  ProgressCallback
	QueueInfo QueueInfoCallback // Called when queue info is available (for recovery)
}

// GetProgress returns the progress callback
func (r *BaseVideoRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// GetQueueInfo returns the queue info callback
func (r *BaseVideoRequest) GetQueueInfo() QueueInfoCallback {
	return r.QueueInfo
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

// QueueInfoCallback is called when queue info is available, before polling starts
// This allows storing queue info for recovery purposes
type QueueInfoCallback func(queueID, responseURL string)

// Error represents a Fal.ai API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// HiDreamOptions represents common options for fal-ai/hidream models
type HiDreamOptions struct {
	NegativePrompt      string   `json:"negative_prompt,omitempty"`       // Default: ""
	ImageSize           string   `json:"image_size,omitempty"`            // Default: "square_hd"
	NumInferenceSteps   *int     `json:"num_inference_steps,omitempty"`   // Default: 50
	Seed                *int     `json:"seed,omitempty"`                  // Optional seed
	GuidanceScale       *float64 `json:"guidance_scale,omitempty"`        // Default: 5.0
	SyncMode            bool     `json:"sync_mode,omitempty"`             // Default: false
	NumImages           int      `json:"num_images,omitempty"`            // Default: 1
	EnableSafetyChecker *bool    `json:"enable_safety_checker,omitempty"` // Default: true
	OutputFormat        string   `json:"output_format,omitempty"`         // Default: "jpeg"
}

// GetDefaultValues returns default values for HiDream options
func (o *HiDreamOptions) GetDefaultValues() map[string]interface{} {
	defaultNumSteps := 50
	defaultGuidanceScale := 5.0
	defaultSafetyChecker := true
	return map[string]interface{}{
		"negative_prompt":       "",
		"image_size":            "square_hd", // Defaulting to 1024x1024
		"num_inference_steps":   &defaultNumSteps,
		"guidance_scale":        &defaultGuidanceScale,
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker,
		"output_format":         "jpeg",
	}
}

// Validate validates HiDream options
func (o *HiDreamOptions) Validate() error {
	validImageSizes := map[string]bool{
		"square_hd": true, "square": true, "portrait_4_3": true,
		"portrait_16_9": true, "landscape_4_3": true, "landscape_16_9": true,
	}
	validOutputFormats := map[string]bool{"jpeg": true, "png": true, "": true}

	if o.ImageSize != "" && !validImageSizes[o.ImageSize] {
		return fmt.Errorf("invalid image_size: %s", o.ImageSize)
	}
	if o.NumInferenceSteps != nil && *o.NumInferenceSteps <= 0 {
		return fmt.Errorf("num_inference_steps must be positive: %d", *o.NumInferenceSteps)
	}
	if o.GuidanceScale != nil && *o.GuidanceScale < 0 {
		return fmt.Errorf("guidance_scale cannot be negative: %f", *o.GuidanceScale)
	}
	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}
	if o.OutputFormat != "" && !validOutputFormats[o.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s", o.OutputFormat)
	}
	return nil
}

// HiDreamI1FullRequest represents a request for the fal-ai/hidream-i1-full model
type HiDreamI1FullRequest struct {
	BaseImageRequest
	NegativePrompt      string   `json:"negative_prompt,omitempty"`
	ImageSize           string   `json:"image_size,omitempty"`
	NumInferenceSteps   *int     `json:"num_inference_steps,omitempty"`
	Seed                *int     `json:"seed,omitempty"`
	GuidanceScale       *float64 `json:"guidance_scale,omitempty"`
	SyncMode            bool     `json:"sync_mode,omitempty"`
	NumImages           int      `json:"num_images,omitempty"`
	EnableSafetyChecker *bool    `json:"enable_safety_checker,omitempty"`
	OutputFormat        string   `json:"output_format,omitempty"`
}

// HiDreamI1DevRequest represents a request for the fal-ai/hidream-i1-dev model
type HiDreamI1DevRequest struct {
	HiDreamI1FullRequest // Assuming same params as full
}

// HiDreamI1FastRequest represents a request for the fal-ai/hidream-i1-fast model
type HiDreamI1FastRequest struct {
	HiDreamI1FullRequest // Assuming same params as full
}

// FluxProV1_1UltraOptions represents options for fal-ai/flux-pro/v1.1-ultra
type FluxProV1_1UltraOptions struct {
	Seed                *int   `json:"seed,omitempty"`
	SyncMode            bool   `json:"sync_mode,omitempty"`
	NumImages           int    `json:"num_images,omitempty"`
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`
	SafetyTolerance     string `json:"safety_tolerance,omitempty"`
	OutputFormat        string `json:"output_format,omitempty"`
	AspectRatio         string `json:"aspect_ratio,omitempty"` // Different from image_size
	Raw                 *bool  `json:"raw,omitempty"`
}

// GetDefaultValues returns default values for Flux Pro v1.1 Ultra options
func (o *FluxProV1_1UltraOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	defaultRaw := false
	return map[string]interface{}{
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker,
		"safety_tolerance":      "2",
		"output_format":         "jpeg",
		"aspect_ratio":          "16:9",
		"raw":                   &defaultRaw,
	}
}

// Validate validates Flux Pro v1.1 Ultra options
func (o *FluxProV1_1UltraOptions) Validate() error {
	validSafetyTolerances := map[string]bool{"1": true, "2": true, "3": true, "4": true, "5": true, "6": true, "": true}
	validOutputFormats := map[string]bool{"jpeg": true, "png": true, "": true}
	validAspectRatios := map[string]bool{
		"21:9": true, "16:9": true, "4:3": true, "3:2": true, "1:1": true,
		"2:3": true, "3:4": true, "9:16": true, "9:21": true, "": true,
	}

	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}
	if o.SafetyTolerance != "" && !validSafetyTolerances[o.SafetyTolerance] {
		return fmt.Errorf("invalid safety_tolerance: %s (must be 1-6)", o.SafetyTolerance)
	}
	if o.OutputFormat != "" && !validOutputFormats[o.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s (must be jpeg or png)", o.OutputFormat)
	}
	if o.AspectRatio != "" && !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s", o.AspectRatio)
	}
	return nil
}

// FluxDevOptions represents the options available for the fal-ai/flux/dev model
type FluxDevOptions struct {
	ImageSize           string  `json:"image_size,omitempty"`            // square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`   // Default: 28
	Seed                *int    `json:"seed,omitempty"`                  // Optional seed
	GuidanceScale       float64 `json:"guidance_scale,omitempty"`        // Default: 3.5
	SyncMode            bool    `json:"sync_mode,omitempty"`             // Default: false
	NumImages           int     `json:"num_images,omitempty"`            // Default: 1
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"` // Default: true
	OutputFormat        string  `json:"output_format,omitempty"`         // jpeg, png. Default: jpeg
}

// GetDefaultValues returns the default values for Flux Dev options
func (o *FluxDevOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	return map[string]interface{}{
		"image_size":            "landscape_4_3",
		"num_inference_steps":   28,
		"guidance_scale":        3.5,
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker,
		"output_format":         "jpeg",
	}
}

// Validate validates the Flux Dev options
func (o *FluxDevOptions) Validate() error {
	validImageSizes := map[string]bool{
		"square_hd":      true,
		"square":         true,
		"portrait_4_3":   true,
		"portrait_16_9":  true,
		"landscape_4_3":  true,
		"landscape_16_9": true,
	}

	if o.ImageSize != "" && !validImageSizes[o.ImageSize] {
		return fmt.Errorf("invalid image_size: %s (must be one of: square_hd, square, portrait_4_3, portrait_16_9, landscape_4_3, landscape_16_9)", o.ImageSize)
	}
	if o.NumInferenceSteps < 0 {
		return fmt.Errorf("num_inference_steps cannot be negative: %d", o.NumInferenceSteps)
	}
	if o.GuidanceScale < 0 {
		return fmt.Errorf("guidance_scale cannot be negative: %f", o.GuidanceScale)
	}
	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}

	return nil
}

// FluxDevRequest represents a request to generate an image using fal-ai/flux/dev
type FluxDevRequest struct {
	BaseImageRequest
	ImageSize           string  `json:"image_size,omitempty"`
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`
	Seed                *int    `json:"seed,omitempty"`
	GuidanceScale       float64 `json:"guidance_scale,omitempty"`
	SyncMode            bool    `json:"sync_mode,omitempty"`
	NumImages           int     `json:"num_images,omitempty"`
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"`
	OutputFormat        string  `json:"output_format,omitempty"`
}

// StableDiffusionV35LargeOptions represents options for fal-ai/stable-diffusion-v35-large
type StableDiffusionV35LargeOptions struct {
	ImageSize           string   `json:"image_size,omitempty"`            // Default: square_hd
	NumInferenceSteps   int      `json:"num_inference_steps,omitempty"`   // Default: 40
	Seed                *int     `json:"seed,omitempty"`                  // Optional
	GuidanceScale       float64  `json:"guidance_scale,omitempty"`        // Default: 4.5
	NumImages           int      `json:"num_images,omitempty"`            // Default: 1
	EnableSafetyChecker *bool    `json:"enable_safety_checker,omitempty"` // Default: true
	OutputFormat        string   `json:"output_format,omitempty"`         // jpeg, png. Default: jpeg
	NegativePrompt      string   `json:"negative_prompt,omitempty"`       // Optional
	PromptExpansion     *bool    `json:"prompt_expansion,omitempty"`      // Default: true
}

// GetDefaultValues returns the default values for Stable Diffusion 3.5 Large options
func (o *StableDiffusionV35LargeOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	defaultPromptExpansion := true
	return map[string]interface{}{
		"image_size":            "square_hd",
		"num_inference_steps":   40,
		"guidance_scale":        4.5,
		"num_images":            1,
		"enable_safety_checker": &defaultSafetyChecker,
		"output_format":         "jpeg",
		"negative_prompt":       "",
		"prompt_expansion":      &defaultPromptExpansion,
	}
}

// Validate validates Stable Diffusion 3.5 Large options
func (o *StableDiffusionV35LargeOptions) Validate() error {
	validImageSizes := map[string]bool{
		"square_hd":      true,
		"square":         true,
		"portrait_4_3":   true,
		"portrait_16_9":  true,
		"landscape_4_3":  true,
		"landscape_16_9": true,
	}
	validOutputFormats := map[string]bool{"jpeg": true, "png": true, "": true}

	if o.ImageSize != "" && !validImageSizes[o.ImageSize] {
		return fmt.Errorf("invalid image_size: %s", o.ImageSize)
	}
	if o.NumInferenceSteps < 0 {
		return fmt.Errorf("num_inference_steps cannot be negative: %d", o.NumInferenceSteps)
	}
	if o.GuidanceScale < 0 {
		return fmt.Errorf("guidance_scale cannot be negative: %f", o.GuidanceScale)
	}
	if o.NumImages < 0 {
		return fmt.Errorf("num_images cannot be negative: %d", o.NumImages)
	}
	if o.OutputFormat != "" && !validOutputFormats[o.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s (must be jpeg or png)", o.OutputFormat)
	}
	return nil
}

// StableDiffusionV35LargeRequest represents a request for fal-ai/stable-diffusion-v35-large
type StableDiffusionV35LargeRequest struct {
	BaseImageRequest
	ImageSize           string  `json:"image_size,omitempty"`
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`
	Seed                *int    `json:"seed,omitempty"`
	GuidanceScale       float64 `json:"guidance_scale,omitempty"`
	NumImages           int     `json:"num_images,omitempty"`
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"`
	OutputFormat        string  `json:"output_format,omitempty"`
	NegativePrompt      string  `json:"negative_prompt,omitempty"`
	PromptExpansion     *bool   `json:"prompt_expansion,omitempty"`
}

// CartoonifyOptions represents options for the cartoonify model.
type CartoonifyOptions struct {
	// No specific options identified yet
}

func (o *CartoonifyOptions) GetDefaultValues() map[string]interface{} {
	return make(map[string]interface{})
}
func (o *CartoonifyOptions) Validate() error { return nil }

// StarVectorOptions represents options for the star-vector model.
type StarVectorOptions struct {
	// No specific options identified yet
}

func (o *StarVectorOptions) GetDefaultValues() map[string]interface{} {
	return make(map[string]interface{})
}
func (o *StarVectorOptions) Validate() error { return nil }

// FluxProV1_1UltraRequest represents a request for fal-ai/flux-pro/v1.1-ultra
type FluxProV1_1UltraRequest struct {
	BaseImageRequest
	Seed                *int   `json:"seed,omitempty"`
	SyncMode            bool   `json:"sync_mode,omitempty"`
	NumImages           int    `json:"num_images,omitempty"`
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`
	SafetyTolerance     string `json:"safety_tolerance,omitempty"`
	OutputFormat        string `json:"output_format,omitempty"`
	AspectRatio         string `json:"aspect_ratio,omitempty"`
	Raw                 *bool  `json:"raw,omitempty"`
}

// CartoonifyRequest represents a request for the cartoonify model
type CartoonifyRequest struct {
	BaseImageRequest // Requires ImageURL
}

// StarVectorRequest represents a request for the star-vector model
type StarVectorRequest struct {
	BaseImageRequest // Requires ImageURL
}

// MinimaxDirectorRequest represents a request to generate a video using the minimax-video-01-director model
type MinimaxDirectorRequest struct {
	BaseVideoRequest
	PromptOptimizer *bool `json:"prompt_optimizer,omitempty"`
}

// MinimaxSubjectReferenceOptions represents options for minimax/video-01-subject-reference
type MinimaxSubjectReferenceOptions struct {
	PromptOptimizer *bool `json:"prompt_optimizer,omitempty"` // Default: true
}

// GetDefaultValues returns default values for MinimaxSubjectReferenceOptions
func (o *MinimaxSubjectReferenceOptions) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"prompt_optimizer": &defaultOptimizer,
	}
}

// Validate validates MinimaxSubjectReferenceOptions
func (o *MinimaxSubjectReferenceOptions) Validate() error {
	// No validation needed for a boolean flag yet
	return nil
}

// MinimaxSubjectReferenceRequest represents a request for minimax/video-01-subject-reference
type MinimaxSubjectReferenceRequest struct {
	BaseVideoRequest                // Embeds Prompt, Progress, Model, Options
	SubjectReferenceImageURL string `json:"subject_reference_image_url"` // Specific required field
	PromptOptimizer          *bool  `json:"prompt_optimizer,omitempty"`  // Mirrored option
}

// MinimaxLiveOptions represents options for minimax/video-01-live
type MinimaxLiveOptions struct {
	PromptOptimizer *bool `json:"prompt_optimizer,omitempty"` // Default: true
}

// GetDefaultValues returns default values for MinimaxLiveOptions
func (o *MinimaxLiveOptions) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"prompt_optimizer": &defaultOptimizer,
	}
}

// Validate validates MinimaxLiveOptions
func (o *MinimaxLiveOptions) Validate() error {
	// No validation needed for a boolean flag yet
	return nil
}

// MinimaxLiveRequest represents a request for minimax/video-01-live
type MinimaxLiveRequest struct {
	BaseVideoRequest       // Embeds Prompt, ImageURL, Progress, Model, Options
	PromptOptimizer  *bool `json:"prompt_optimizer,omitempty"` // Mirrored option
}

// MinimaxVideo01Options represents options for minimax/video-01
type MinimaxVideo01Options struct {
	PromptOptimizer *bool `json:"prompt_optimizer,omitempty"` // Default: true
}

// GetDefaultValues returns default values for MinimaxVideo01Options
func (o *MinimaxVideo01Options) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"prompt_optimizer": &defaultOptimizer,
	}
}

// Validate validates MinimaxVideo01Options
func (o *MinimaxVideo01Options) Validate() error {
	// No validation needed for a boolean flag yet
	return nil
}

// MinimaxVideo01Request represents a request for minimax/video-01
type MinimaxVideo01Request struct {
	BaseVideoRequest       // Embeds Prompt, Progress, Model, Options (ImageURL should be empty)
	PromptOptimizer  *bool `json:"prompt_optimizer,omitempty"` // Mirrored option
}

// MiniMax Hailuo-02 Text To Video request
type MinimaxHailuo02Request struct {
	BaseVideoRequest
	Duration        string `json:"duration,omitempty"`
	PromptOptimizer *bool  `json:"prompt_optimizer,omitempty"`
}

// Veo3Options represents options for the fal-ai/veo3/image-to-video model
type Veo3Options struct {
	AspectRatio   string `json:"aspect_ratio,omitempty"`   // auto, 16:9, 9:16
	Duration      string `json:"duration,omitempty"`       // 4s, 6s, 8s
	Resolution    string `json:"resolution,omitempty"`     // 720p, 1080p
	GenerateAudio *bool  `json:"generate_audio,omitempty"` // default: true
	AutoFix       *bool  `json:"auto_fix,omitempty"`       // auto-fix failed prompts
}

// GetDefaultValues returns default values for Veo3 options
func (o *Veo3Options) GetDefaultValues() map[string]interface{} {
	defaultAudio := true
	defaultAutoFix := false
	return map[string]interface{}{
		"aspect_ratio":   "16:9",
		"duration":       "8s",
		"resolution":     "720p",
		"generate_audio": &defaultAudio,
		"auto_fix":       &defaultAutoFix,
	}
}

// Validate validates Veo3 options
func (o *Veo3Options) Validate() error {
	validAspectRatios := map[string]bool{
		"auto": true,
		"16:9": true,
		"9:16": true,
		"":     true,
	}
	validDurations := map[string]bool{
		"4s": true,
		"6s": true,
		"8s": true,
		"":   true,
	}
	validResolutions := map[string]bool{
		"720p":  true,
		"1080p": true,
		"":      true,
	}

	if !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s (must be one of: auto, 16:9, 9:16)", o.AspectRatio)
	}
	if !validDurations[o.Duration] {
		return fmt.Errorf("invalid duration: %s (must be one of: 4s, 6s, 8s)", o.Duration)
	}
	if !validResolutions[o.Resolution] {
		return fmt.Errorf("invalid resolution: %s (must be one of: 720p, 1080p)", o.Resolution)
	}
	return nil
}

// Veo3Request represents a request for fal-ai/veo3/image-to-video
type Veo3Request struct {
	BaseVideoRequest
	AspectRatio   string `json:"aspect_ratio,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Resolution    string `json:"resolution,omitempty"`
	GenerateAudio *bool  `json:"generate_audio,omitempty"`
	AutoFix       *bool  `json:"auto_fix,omitempty"`
}

// Veo31FastOptions represents options for fal-ai/veo3.1/fast/image-to-video
type Veo31FastOptions struct {
	AspectRatio   string `json:"aspect_ratio,omitempty"`   // auto, 16:9, 9:16 (default: auto)
	Duration      string `json:"duration,omitempty"`       // 4s, 6s, 8s (default: 8s)
	Resolution    string `json:"resolution,omitempty"`     // 720p, 1080p (default: 720p)
	GenerateAudio *bool  `json:"generate_audio,omitempty"` // default: true
	AutoFix       *bool  `json:"auto_fix,omitempty"`       // auto-fix failed prompts
}

// GetDefaultValues returns default values for Veo31Fast options
func (o *Veo31FastOptions) GetDefaultValues() map[string]interface{} {
	defaultAudio := true
	defaultAutoFix := false
	return map[string]interface{}{
		"aspect_ratio":   "auto",
		"duration":       "8s",
		"resolution":     "720p",
		"generate_audio": &defaultAudio,
		"auto_fix":       &defaultAutoFix,
	}
}

// Validate validates Veo31Fast options
func (o *Veo31FastOptions) Validate() error {
	validAspectRatios := map[string]bool{
		"auto": true,
		"16:9": true,
		"9:16": true,
		"":     true,
	}
	validDurations := map[string]bool{
		"4s": true,
		"6s": true,
		"8s": true,
		"":   true,
	}
	validResolutions := map[string]bool{
		"720p":  true,
		"1080p": true,
		"":      true,
	}

	if !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s (must be one of: auto, 16:9, 9:16)", o.AspectRatio)
	}
	if !validDurations[o.Duration] {
		return fmt.Errorf("invalid duration: %s (must be one of: 4s, 6s, 8s)", o.Duration)
	}
	if !validResolutions[o.Resolution] {
		return fmt.Errorf("invalid resolution: %s (must be one of: 720p, 1080p)", o.Resolution)
	}
	return nil
}

// Veo31FastRequest represents a request for fal-ai/veo3.1/fast/image-to-video
type Veo31FastRequest struct {
	BaseVideoRequest
	AspectRatio   string `json:"aspect_ratio,omitempty"`
	Duration      string `json:"duration,omitempty"`
	Resolution    string `json:"resolution,omitempty"`
	GenerateAudio *bool  `json:"generate_audio,omitempty"`
	AutoFix       *bool  `json:"auto_fix,omitempty"`
}

// HunyuanVideoOptions represents options for fal-ai/hunyuan-video
type HunyuanVideoOptions struct {
	AspectRatio         string `json:"aspect_ratio,omitempty"`           // 16:9, 9:16, 4:3, 3:4, 1:1. Default: 16:9
	Resolution          string `json:"resolution,omitempty"`             // 480p, 580p, 720p, 1080p. Default: 720p
	VideoLength         string `json:"video_length,omitempty"`           // 5s, 10s. Default: 5s
	NumInferenceSteps   int    `json:"num_inference_steps,omitempty"`    // Default: 50
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`  // Default: true
	Seed                *int   `json:"seed,omitempty"`                   // Optional
	EmbeddedGuidance    *bool  `json:"embedded_guidance_scale,omitempty"` // Default: 6.0
}

// GetDefaultValues returns the default values for Hunyuan Video options
func (o *HunyuanVideoOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	return map[string]interface{}{
		"aspect_ratio":          "16:9",
		"resolution":            "720p",
		"video_length":          "5s",
		"num_inference_steps":   50,
		"enable_safety_checker": &defaultSafetyChecker,
	}
}

// Validate validates Hunyuan Video options
func (o *HunyuanVideoOptions) Validate() error {
	validAspectRatios := map[string]bool{
		"16:9": true, "9:16": true, "4:3": true, "3:4": true, "1:1": true, "": true,
	}
	validResolutions := map[string]bool{
		"480p": true, "580p": true, "720p": true, "1080p": true, "": true,
	}
	validVideoLengths := map[string]bool{
		"5s": true, "10s": true, "": true,
	}

	if !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s", o.AspectRatio)
	}
	if !validResolutions[o.Resolution] {
		return fmt.Errorf("invalid resolution: %s", o.Resolution)
	}
	if !validVideoLengths[o.VideoLength] {
		return fmt.Errorf("invalid video_length: %s", o.VideoLength)
	}
	return nil
}

// HunyuanVideoRequest represents a request for fal-ai/hunyuan-video
type HunyuanVideoRequest struct {
	BaseVideoRequest
	AspectRatio         string `json:"aspect_ratio,omitempty"`
	Resolution          string `json:"resolution,omitempty"`
	VideoLength         string `json:"video_length,omitempty"`
	NumInferenceSteps   int    `json:"num_inference_steps,omitempty"`
	EnableSafetyChecker *bool  `json:"enable_safety_checker,omitempty"`
}

// KlingVideoV25Options represents options for fal-ai/kling-video/v2.5-turbo/pro models
type KlingVideoV25Options struct {
	Duration       string  `json:"duration,omitempty"`        // 5, 10. Default: 5
	AspectRatio    string  `json:"aspect_ratio,omitempty"`    // 16:9, 9:16, 1:1. Default: 16:9
	NegativePrompt string  `json:"negative_prompt,omitempty"` // Default: blur, distort, low quality
	CFGScale       float64 `json:"cfg_scale,omitempty"`       // 0-1. Default: 0.5
}

// GetDefaultValues returns the default values for Kling Video v2.5 options
func (o *KlingVideoV25Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"duration":        "5",
		"aspect_ratio":    "16:9",
		"negative_prompt": "blur, distort, and low quality",
		"cfg_scale":       0.5,
	}
}

// Validate validates Kling Video v2.5 options
func (o *KlingVideoV25Options) Validate() error {
	validDurations := map[string]bool{"5": true, "10": true, "": true}
	validAspectRatios := map[string]bool{"16:9": true, "9:16": true, "1:1": true, "": true}

	if !validDurations[o.Duration] {
		return fmt.Errorf("invalid duration: %s (must be 5 or 10)", o.Duration)
	}
	if !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s", o.AspectRatio)
	}
	if o.CFGScale < 0 || o.CFGScale > 1 {
		return fmt.Errorf("invalid cfg_scale: %f (must be between 0 and 1)", o.CFGScale)
	}
	return nil
}

// KlingVideoV25Request represents a request for Kling Video v2.5 models
type KlingVideoV25Request struct {
	BaseVideoRequest
	Duration       string  `json:"duration,omitempty"`
	AspectRatio    string  `json:"aspect_ratio,omitempty"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	CFGScale       float64 `json:"cfg_scale,omitempty"`
}

// LumaDreamMachineOptions represents options for fal-ai/luma-dream-machine
type LumaDreamMachineOptions struct {
	AspectRatio string `json:"aspect_ratio,omitempty"` // 16:9, 9:16, 4:3, 3:4, 21:9, 9:21, 1:1. Default: 16:9
	Loop        *bool  `json:"loop,omitempty"`         // Default: false
}

// GetDefaultValues returns the default values for Luma Dream Machine options
func (o *LumaDreamMachineOptions) GetDefaultValues() map[string]interface{} {
	defaultLoop := false
	return map[string]interface{}{
		"aspect_ratio": "16:9",
		"loop":         &defaultLoop,
	}
}

// Validate validates Luma Dream Machine options
func (o *LumaDreamMachineOptions) Validate() error {
	validAspectRatios := map[string]bool{
		"16:9": true, "9:16": true, "4:3": true, "3:4": true,
		"21:9": true, "9:21": true, "1:1": true, "": true,
	}
	if !validAspectRatios[o.AspectRatio] {
		return fmt.Errorf("invalid aspect_ratio: %s", o.AspectRatio)
	}
	return nil
}

// LumaDreamMachineRequest represents a request for fal-ai/luma-dream-machine
type LumaDreamMachineRequest struct {
	BaseVideoRequest
	AspectRatio string `json:"aspect_ratio,omitempty"`
	Loop        *bool  `json:"loop,omitempty"`
}

// LTXVideo13BOptions represents options for fal-ai/ltx-video-13b-distilled/multiconditioning
type LTXVideo13BOptions struct {
	NumFrames           int     `json:"num_frames,omitempty"`            // Default: 97
	FrameRate           int     `json:"frame_rate,omitempty"`            // Default: 25
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`   // Default: 30
	GuidanceScale       float64 `json:"guidance_scale,omitempty"`        // Default: 3.0
	Seed                *int    `json:"seed,omitempty"`                  // Optional
	NegativePrompt      string  `json:"negative_prompt,omitempty"`       // Optional
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"` // Default: true
}

// GetDefaultValues returns the default values for LTX Video 13B options
func (o *LTXVideo13BOptions) GetDefaultValues() map[string]interface{} {
	defaultSafetyChecker := true
	return map[string]interface{}{
		"num_frames":            97,
		"frame_rate":            25,
		"num_inference_steps":   30,
		"guidance_scale":        3.0,
		"enable_safety_checker": &defaultSafetyChecker,
		"negative_prompt":       "",
	}
}

// Validate validates LTX Video 13B options
func (o *LTXVideo13BOptions) Validate() error {
	if o.NumFrames < 0 {
		return fmt.Errorf("num_frames cannot be negative: %d", o.NumFrames)
	}
	if o.FrameRate < 0 {
		return fmt.Errorf("frame_rate cannot be negative: %d", o.FrameRate)
	}
	if o.NumInferenceSteps < 0 {
		return fmt.Errorf("num_inference_steps cannot be negative: %d", o.NumInferenceSteps)
	}
	if o.GuidanceScale < 0 {
		return fmt.Errorf("guidance_scale cannot be negative: %f", o.GuidanceScale)
	}
	return nil
}

// LTXVideo13BRequest represents a request for fal-ai/ltx-video-13b-distilled/multiconditioning
type LTXVideo13BRequest struct {
	BaseVideoRequest
	NumFrames           int     `json:"num_frames,omitempty"`
	FrameRate           int     `json:"frame_rate,omitempty"`
	NumInferenceSteps   int     `json:"num_inference_steps,omitempty"`
	GuidanceScale       float64 `json:"guidance_scale,omitempty"`
	NegativePrompt      string  `json:"negative_prompt,omitempty"`
	EnableSafetyChecker *bool   `json:"enable_safety_checker,omitempty"`
}

// TopazUpscaleVideoOptions represents options for fal-ai/topaz/upscale/video
type TopazUpscaleVideoOptions struct {
	Model      string `json:"model,omitempty"`       // Default: auto
	OutputType string `json:"output_type,omitempty"` // mp4, mov. Default: mp4
}

// GetDefaultValues returns the default values for Topaz Upscale Video options
func (o *TopazUpscaleVideoOptions) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"model":       "auto",
		"output_type": "mp4",
	}
}

// Validate validates Topaz Upscale Video options
func (o *TopazUpscaleVideoOptions) Validate() error {
	validOutputTypes := map[string]bool{"mp4": true, "mov": true, "": true}
	if !validOutputTypes[o.OutputType] {
		return fmt.Errorf("invalid output_type: %s (must be mp4 or mov)", o.OutputType)
	}
	return nil
}

// TopazUpscaleVideoRequest represents a request for fal-ai/topaz/upscale/video
type TopazUpscaleVideoRequest struct {
	VideoURL   string `json:"video_url"`
	Model      string `json:"model,omitempty"`
	OutputType string `json:"output_type,omitempty"`
	Progress   ProgressCallback
}

// GetProgress returns the progress callback
func (r *TopazUpscaleVideoRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// SyncLipsyncV2Options represents options for fal-ai/sync-lipsync/v2
type SyncLipsyncV2Options struct {
	Model      string `json:"model,omitempty"`       // wav2lip, wav2lip_gan. Default: wav2lip
	OutputType string `json:"output_type,omitempty"` // mp4, webm. Default: mp4
}

// GetDefaultValues returns the default values for Sync Lipsync V2 options
func (o *SyncLipsyncV2Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"model":       "wav2lip",
		"output_type": "mp4",
	}
}

// Validate validates Sync Lipsync V2 options
func (o *SyncLipsyncV2Options) Validate() error {
	validModels := map[string]bool{"wav2lip": true, "wav2lip_gan": true, "": true}
	validOutputTypes := map[string]bool{"mp4": true, "webm": true, "": true}
	if !validModels[o.Model] {
		return fmt.Errorf("invalid model: %s", o.Model)
	}
	if !validOutputTypes[o.OutputType] {
		return fmt.Errorf("invalid output_type: %s", o.OutputType)
	}
	return nil
}

// SyncLipsyncV2Request represents a request for fal-ai/sync-lipsync/v2
type SyncLipsyncV2Request struct {
	VideoURL   string `json:"video_url"`
	AudioURL   string `json:"audio_url"`
	Model      string `json:"model,omitempty"`
	OutputType string `json:"output_type,omitempty"`
	Progress   ProgressCallback
}

// GetProgress returns the progress callback
func (r *SyncLipsyncV2Request) GetProgress() ProgressCallback {
	return r.Progress
}

// MMAudioV2Options represents options for fal-ai/mmaudio-v2
type MMAudioV2Options struct {
	Duration         float64 `json:"duration,omitempty"`          // Output duration. Default: video duration
	NumInferenceSteps int    `json:"num_inference_steps,omitempty"` // Default: 25
	Seed             *int    `json:"seed,omitempty"`               // Optional
}

// GetDefaultValues returns the default values for MMAudio V2 options
func (o *MMAudioV2Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"num_inference_steps": 25,
	}
}

// Validate validates MMAudio V2 options
func (o *MMAudioV2Options) Validate() error {
	if o.Duration < 0 {
		return fmt.Errorf("duration cannot be negative: %f", o.Duration)
	}
	if o.NumInferenceSteps < 0 {
		return fmt.Errorf("num_inference_steps cannot be negative: %d", o.NumInferenceSteps)
	}
	return nil
}

// MMAudioV2Request represents a request for fal-ai/mmaudio-v2
type MMAudioV2Request struct {
	VideoURL          string  `json:"video_url"`
	Prompt            string  `json:"prompt,omitempty"`
	Duration          float64 `json:"duration,omitempty"`
	NumInferenceSteps int     `json:"num_inference_steps,omitempty"`
	Progress          ProgressCallback
}

// GetProgress returns the progress callback
func (r *MMAudioV2Request) GetProgress() ProgressCallback {
	return r.Progress
}

// MinimaxMusicV2Options represents options for fal-ai/minimax-music/v2
type MinimaxMusicV2Options struct {
	Duration     int    `json:"duration,omitempty"`      // 1-300 seconds. Default: 60
	ReferenceAudioURL string `json:"reference_audio_url,omitempty"` // Optional reference audio
}

// GetDefaultValues returns the default values for Minimax Music V2 options
func (o *MinimaxMusicV2Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"duration": 60,
	}
}

// Validate validates Minimax Music V2 options
func (o *MinimaxMusicV2Options) Validate() error {
	if o.Duration < 1 || o.Duration > 300 {
		return fmt.Errorf("duration must be between 1 and 300 seconds: %d", o.Duration)
	}
	return nil
}

// MinimaxMusicV2Request represents a request for fal-ai/minimax-music/v2
type MinimaxMusicV2Request struct {
	Prompt            string `json:"prompt"`
	Duration          int    `json:"duration,omitempty"`
	ReferenceAudioURL string `json:"reference_audio_url,omitempty"`
	Progress          ProgressCallback
}

// GetProgress returns the progress callback
func (r *MinimaxMusicV2Request) GetProgress() ProgressCallback {
	return r.Progress
}

// StableAudio25Options represents options for fal-ai/stable-audio-25/text-to-audio
type StableAudio25Options struct {
	Duration    float64 `json:"duration,omitempty"`     // 1-180 seconds. Default: 30
	SampleRate  int     `json:"sample_rate,omitempty"`  // Default: 44100
	OutputFormat string `json:"output_format,omitempty"` // wav, mp3, ogg. Default: wav
	Seed        *int    `json:"seed,omitempty"`         // Optional
}

// GetDefaultValues returns the default values for Stable Audio 2.5 options
func (o *StableAudio25Options) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"duration":      30.0,
		"sample_rate":   44100,
		"output_format": "wav",
	}
}

// Validate validates Stable Audio 2.5 options
func (o *StableAudio25Options) Validate() error {
	if o.Duration < 1 || o.Duration > 180 {
		return fmt.Errorf("duration must be between 1 and 180 seconds: %f", o.Duration)
	}
	validFormats := map[string]bool{"wav": true, "mp3": true, "ogg": true, "": true}
	if !validFormats[o.OutputFormat] {
		return fmt.Errorf("invalid output_format: %s", o.OutputFormat)
	}
	return nil
}

// StableAudio25Request represents a request for fal-ai/stable-audio-25/text-to-audio
type StableAudio25Request struct {
	Prompt       string  `json:"prompt"`
	Duration     float64 `json:"duration,omitempty"`
	SampleRate   int     `json:"sample_rate,omitempty"`
	OutputFormat string  `json:"output_format,omitempty"`
	Progress     ProgressCallback
}

// GetProgress returns the progress callback
func (r *StableAudio25Request) GetProgress() ProgressCallback {
	return r.Progress
}

// ChatterboxTTSOptions represents options for fal-ai/chatterbox/text-to-speech/turbo
type ChatterboxTTSOptions struct {
	Exaggeration float64 `json:"exaggeration,omitempty"` // 0.0-1.0. Default: 0.5
	CFGWeight    float64 `json:"cfg_weight,omitempty"`   // 0.0-1.0. Default: 0.5
}

// GetDefaultValues returns the default values for Chatterbox TTS options
func (o *ChatterboxTTSOptions) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"exaggeration": 0.5,
		"cfg_weight":   0.5,
	}
}

// Validate validates Chatterbox TTS options
func (o *ChatterboxTTSOptions) Validate() error {
	if o.Exaggeration < 0 || o.Exaggeration > 1 {
		return fmt.Errorf("exaggeration must be between 0 and 1: %f", o.Exaggeration)
	}
	if o.CFGWeight < 0 || o.CFGWeight > 1 {
		return fmt.Errorf("cfg_weight must be between 0 and 1: %f", o.CFGWeight)
	}
	return nil
}

// ChatterboxTTSRequest represents a request for fal-ai/chatterbox/text-to-speech/turbo
type ChatterboxTTSRequest struct {
	Text         string  `json:"text"`
	AudioPrompt  string  `json:"audio_prompt_url,omitempty"` // Optional reference audio
	Exaggeration float64 `json:"exaggeration,omitempty"`
	CFGWeight    float64 `json:"cfg_weight,omitempty"`
	Progress     ProgressCallback
}

// GetProgress returns the progress callback
func (r *ChatterboxTTSRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// ElevenLabsDialogOptions represents options for fal-ai/elevenlabs/text-to-dialogue/eleven-v3
type ElevenLabsDialogOptions struct {
	VoiceID         string  `json:"voice_id,omitempty"`         // Default: Rachel
	OutputFormat    string  `json:"output_format,omitempty"`    // mp3_22050_32, mp3_44100_64, etc. Default: mp3_22050_32
	Stability       float64 `json:"stability,omitempty"`        // 0-1. Default: 0.5
	SimilarityBoost float64 `json:"similarity_boost,omitempty"` // 0-1. Default: 0.75
}

// GetDefaultValues returns the default values for ElevenLabs Dialog options
func (o *ElevenLabsDialogOptions) GetDefaultValues() map[string]interface{} {
	return map[string]interface{}{
		"voice_id":         "Rachel",
		"output_format":    "mp3_22050_32",
		"stability":        0.5,
		"similarity_boost": 0.75,
	}
}

// Validate validates ElevenLabs Dialog options
func (o *ElevenLabsDialogOptions) Validate() error {
	if o.Stability < 0 || o.Stability > 1 {
		return fmt.Errorf("stability must be between 0 and 1: %f", o.Stability)
	}
	if o.SimilarityBoost < 0 || o.SimilarityBoost > 1 {
		return fmt.Errorf("similarity_boost must be between 0 and 1: %f", o.SimilarityBoost)
	}
	return nil
}

// ElevenLabsDialogRequest represents a request for fal-ai/elevenlabs/text-to-dialogue/eleven-v3
type ElevenLabsDialogRequest struct {
	Text            string  `json:"text"`
	VoiceID         string  `json:"voice_id,omitempty"`
	OutputFormat    string  `json:"output_format,omitempty"`
	Stability       float64 `json:"stability,omitempty"`
	SimilarityBoost float64 `json:"similarity_boost,omitempty"`
	Progress        ProgressCallback
}

// GetProgress returns the progress callback
func (r *ElevenLabsDialogRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// ElevenLabsTTSOptions represents options for fal-ai/elevenlabs/tts/turbo-v2.5
type ElevenLabsTTSOptions struct {
	Voice           string   `json:"voice,omitempty"`            // Default: Rachel
	Stability       *float64 `json:"stability,omitempty"`        // 0-1. Default: 0.5
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"` // 0-1. Default: 0.75
	Style           *float64 `json:"style,omitempty"`            // 0-1. Default: 0.0
	Speed           *float64 `json:"speed,omitempty"`            // 0.25-4.0. Default: 1.0
	Timestamps      *bool    `json:"timestamps,omitempty"`       // Default: false
	LanguageCode    string   `json:"language_code,omitempty"`    // Optional language code
}

// GetDefaultValues returns the default values for ElevenLabs TTS options
func (o *ElevenLabsTTSOptions) GetDefaultValues() map[string]interface{} {
	defaultStability := 0.5
	defaultSimilarity := 0.75
	defaultStyle := 0.0
	defaultSpeed := 1.0
	defaultTimestamps := false
	return map[string]interface{}{
		"voice":            "Aria",
		"stability":        &defaultStability,
		"similarity_boost": &defaultSimilarity,
		"style":            &defaultStyle,
		"speed":            &defaultSpeed,
		"timestamps":       &defaultTimestamps,
	}
}

// Validate validates ElevenLabs TTS options
func (o *ElevenLabsTTSOptions) Validate() error {
	if o.Stability != nil && (*o.Stability < 0 || *o.Stability > 1) {
		return fmt.Errorf("stability must be between 0 and 1: %f", *o.Stability)
	}
	if o.SimilarityBoost != nil && (*o.SimilarityBoost < 0 || *o.SimilarityBoost > 1) {
		return fmt.Errorf("similarity_boost must be between 0 and 1: %f", *o.SimilarityBoost)
	}
	if o.Style != nil && (*o.Style < 0 || *o.Style > 1) {
		return fmt.Errorf("style must be between 0 and 1: %f", *o.Style)
	}
	if o.Speed != nil && (*o.Speed < 0.25 || *o.Speed > 4.0) {
		return fmt.Errorf("speed must be between 0.25 and 4.0: %f", *o.Speed)
	}
	return nil
}

// ElevenLabsTTSRequest represents a request for fal-ai/elevenlabs/tts/turbo-v2.5
type ElevenLabsTTSRequest struct {
	BaseSpeechRequest
	Voice           string   `json:"voice,omitempty"`
	Stability       *float64 `json:"stability,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
	Timestamps      *bool    `json:"timestamps,omitempty"`
	LanguageCode    string   `json:"language_code,omitempty"`
	PreviousText    string   `json:"previous_text,omitempty"`
	NextText        string   `json:"next_text,omitempty"`
}

// GetProgress returns the progress callback
func (r *ElevenLabsTTSRequest) GetProgress() ProgressCallback {
	return r.Progress
}

// ==================== Scribe V2 (Audio Transcription) ====================

// ScribeV2Options represents options for fal-ai/scribe/v2
type ScribeV2Options struct {
	Task        string `json:"task,omitempty"`         // transcribe (default), translate
	Language    string `json:"language,omitempty"`     // ISO 639-1 code, auto-detected if not specified
	ChunkLevel  string `json:"chunk_level,omitempty"`  // segment (default), word
	Diarize     *bool  `json:"diarize,omitempty"`      // Speaker diarization (default: true)
	NumSpeakers *int   `json:"num_speakers,omitempty"` // Number of speakers (optional)
}

// GetDefaultValues returns default values for Scribe V2 options
func (o *ScribeV2Options) GetDefaultValues() map[string]interface{} {
	defaultDiarize := true
	return map[string]interface{}{
		"task":        "transcribe",
		"chunk_level": "segment",
		"diarize":     &defaultDiarize,
	}
}

// Validate validates Scribe V2 options
func (o *ScribeV2Options) Validate() error {
	validTasks := map[string]bool{"transcribe": true, "translate": true, "": true}
	validChunkLevels := map[string]bool{"segment": true, "word": true, "": true}

	if !validTasks[o.Task] {
		return fmt.Errorf("invalid task: %s (must be transcribe or translate)", o.Task)
	}
	if !validChunkLevels[o.ChunkLevel] {
		return fmt.Errorf("invalid chunk_level: %s (must be segment or word)", o.ChunkLevel)
	}
	if o.NumSpeakers != nil && (*o.NumSpeakers < 1 || *o.NumSpeakers > 50) {
		return fmt.Errorf("invalid num_speakers: %d (must be between 1 and 50)", *o.NumSpeakers)
	}
	return nil
}

// ScribeV2Request represents a request for fal-ai/scribe/v2
type ScribeV2Request struct {
	AudioURL    string           `json:"audio_url"`
	Task        string           `json:"task,omitempty"`
	Language    string           `json:"language,omitempty"`
	ChunkLevel  string           `json:"chunk_level,omitempty"`
	Diarize     *bool            `json:"diarize,omitempty"`
	NumSpeakers *int             `json:"num_speakers,omitempty"`
	Progress    ProgressCallback `json:"-"`
}

// GetProgress returns the progress callback
func (r *ScribeV2Request) GetProgress() ProgressCallback {
	return r.Progress
}

// ScribeWord represents a transcribed word with timing
type ScribeWord struct {
	Text      string  `json:"text"`
	Start     float64 `json:"start"`     // Start time in seconds
	End       float64 `json:"end"`       // End time in seconds
	Type      string  `json:"type"`      // word, spacing, punctuation
	SpeakerID *int    `json:"speaker_id"` // Speaker ID if diarization enabled
}

// ScribeV2Response represents the response from Scribe V2
type ScribeV2Response struct {
	Text                string       `json:"text"`
	LanguageCode        string       `json:"language_code"`
	LanguageProbability float64      `json:"language_probability"`
	Words               []ScribeWord `json:"words"`
}
