package image

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// ImageRequest represents a request to generate an image
type ImageRequest struct {
	Prompt    string
	ImageURL  string // Optional for text2image
	ModelType string // "text2image" or "image2image"
	ModelName string
	Progress  fal.ProgressCallback
	UserNick  string
	UserID    zkidentity.ShortID
	PriceUSD  float64
	NumImages int // Number of images requested (for models that support it)

	// Model-specific options parsed from command args
	ImageSize           string   // e.g., "landscape_4_3"
	Seed                *int     // Optional seed
	NumInferenceSteps   *int     // Optional steps (e.g., flux/schnell)
	EnableSafetyChecker *bool    // Optional override
	SafetyTolerance     string   // Optional tolerance (e.g., flux-pro)
	OutputFormat        string   // Optional format (e.g., flux-pro)
	NegativePrompt      string   // Optional negative prompt (e.g., hidream)
	GuidanceScale       *float64 // Optional guidance scale (e.g., hidream)
	AspectRatio         string   // Optional aspect ratio string (e.g., flux-ultra)
	Raw                 *bool    // Optional raw flag (e.g., flux-ultra)
}

// ImageResult represents the result of an image generation
type ImageResult struct {
	ImageURL string
	Success  bool
	Error    error
}

// IsSuccess checks if the image generation was successful.
func (r *ImageResult) IsSuccess() bool {
	if r == nil {
		return false
	}
	return r.Success
}

// GetError returns the error from the image generation, if any.
func (r *ImageResult) GetError() error {
	if r == nil {
		return nil // Consistent with IsSuccess returning false for nil
	}
	return r.Error
}
