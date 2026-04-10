package faladapter

import (
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
	HelpDoc          string
}
