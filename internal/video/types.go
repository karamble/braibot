package video

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// VideoRequest represents a request to generate a video
type VideoRequest struct {
	Prompt         string
	ImageURL       string // Optional for text2video
	Duration       string
	AspectRatio    string
	NegativePrompt string
	CFGScale       float64
	ModelType      string // "text2video" or "image2video"
	Progress       fal.ProgressCallback
	UserNick       string
	UserID         zkidentity.ShortID
	PriceUSD       float64
}

// VideoOptions represents the options for video generation
type VideoOptions struct {
	Duration       string
	AspectRatio    string
	NegativePrompt string
	CFGScale       float64
}

// VideoResult represents the result of a video generation
type VideoResult struct {
	VideoURL string
	Success  bool
	Error    error
}
