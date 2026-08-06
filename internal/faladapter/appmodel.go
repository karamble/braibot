package faladapter

import (
	"strings"

	"github.com/karamble/braibot/pkg/fal"
)

// AppModel wraps fal.Model with braibot-specific application metadata.
// The embedded fal.Model contains only fal.ai API concerns (Name, Type,
// Endpoint, Options, Description). The additional fields here are
// braibot business logic that does not belong in the standalone fal client.
type AppModel struct {
	fal.Model
	PriceUSD         float64
	PerSecondPricing bool
	// ResolutionPricing overrides the per-second rate by requested
	// resolution; see appModelMeta and RateFor.
	ResolutionPricing map[string]float64
	MaxTextChars      int
	HelpDoc           string
}

// RateFor returns the per-second rate to charge for a resolution. Models
// without ResolutionPricing keep their flat PriceUSD. An unrecognised or empty
// resolution takes the dearest listed rate: guessing low sells below cost, and
// a caller who wants the cheap tier can name it.
func (m AppModel) RateFor(resolution string) float64 {
	if len(m.ResolutionPricing) == 0 {
		return m.PriceUSD
	}
	if r, ok := m.ResolutionPricing[strings.ToLower(strings.TrimSpace(resolution))]; ok {
		return r
	}
	dearest := m.PriceUSD
	for _, r := range m.ResolutionPricing {
		if r > dearest {
			dearest = r
		}
	}
	return dearest
}
