package braibottypes

import (
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/karamble/braibot/pkg/fal"
)

// ExternalBilling reports a charge that was already applied by an external
// biller (the MCP harness) before the service ran. Services with internal
// billing disabled include it in the status DMs instead of reporting billing
// as disabled.
type ExternalBilling struct {
	ChargedDCR float64
	ChargedUSD float64
	BalanceDCR float64 // balance after the charge
}

// GenerationRequest contains fields common to all generation service requests
// (video, image, speech).
type GenerationRequest struct {
	ModelName       string
	ModelType       string
	Progress        fal.ProgressCallback
	UserNick        string
	UserID          zkidentity.ShortID
	PriceUSD        float64
	IsPM            bool   // Whether this is a private message
	GC              string // Group chat name if not PM
	ExternalBilling *ExternalBilling
}
