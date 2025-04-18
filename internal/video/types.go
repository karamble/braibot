package video

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// VideoRequest represents a request to generate a video
type VideoRequest struct {
	Prompt         string
	ImageURL       string   // Optional for text2video
	Duration       string   // Optional, defaults handled by FAL
	AspectRatio    string   // Optional, defaults handled by FAL
	NegativePrompt string   // Optional, defaults handled by FAL
	CFGScale       *float64 // Optional, use pointer to track if set
	ModelType      string   // "text2video" or "image2video"
	Progress       fal.ProgressCallback
	UserNick       string
	UserID         zkidentity.ShortID
	PriceUSD       float64
}

// VideoResult represents the result of a video generation
type VideoResult struct {
	VideoURL string
	Success  bool
	Error    error
}
