package braibottypes

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// GenerationRequest contains fields common to all generation service requests
// (video, image, speech).
type GenerationRequest struct {
	ModelName string
	ModelType string
	Progress  fal.ProgressCallback
	UserNick  string
	UserID    zkidentity.ShortID
	PriceUSD  float64
	IsPM      bool   // Whether this is a private message
	GC        string // Group chat name if not PM
}
