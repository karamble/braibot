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
}

// ImageResult represents the result of an image generation
type ImageResult struct {
	ImageURL string
	Success  bool
	Error    error
}
