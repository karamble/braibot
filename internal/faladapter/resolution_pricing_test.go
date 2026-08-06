// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package faladapter

import "testing"

// upstreamPerSecond is fal's seedance cost at each resolution. 720p and 1080p
// are quoted on the model page; 480p and 4K come from the documented token
// formula (h*w*duration*24/1024 at $0.014 per 1k tokens, $0.008 at 4K), which
// reproduces both quoted figures to the cent.
var upstreamPerSecond = map[string]float64{
	"480p": 0.1345, "720p": 0.3034, "1080p": 0.682, "4k": 1.5552,
}

// TestSeedanceClearsCostAtEveryResolution is the regression for the flat-rate
// defect: seedance is token-billed upstream, so cost scales with pixel count
// while a single PriceUSD stayed flat and lost money at 1080p and 4K.
func TestSeedanceClearsCostAtEveryResolution(t *testing.T) {
	for _, name := range []string{
		"seedance-2.0-text", "seedance-2.0-image",
		"seedance-2.0-fast-image", "seedance-2.0-reference",
	} {
		meta, ok := modelMeta[name]
		if !ok {
			t.Errorf("%s missing from the registry", name)
			continue
		}
		if len(meta.ResolutionPricing) == 0 {
			t.Errorf("%s has no ResolutionPricing - it is token-billed upstream", name)
			continue
		}
		m := AppModel{
			PriceUSD:          meta.PriceUSD,
			PerSecondPricing:  meta.PerSecondPricing,
			ResolutionPricing: meta.ResolutionPricing,
		}
		// The fast variant buys the cheaper upstream tier; scale accordingly.
		discount := 1.0
		if name == "seedance-2.0-fast-image" {
			discount = 0.2419 / 0.3034
		}
		for res, up := range upstreamPerSecond {
			cost := up * discount
			if got := m.RateFor(res); got <= cost {
				t.Errorf("%s at %s: charges $%.4f/s against $%.4f/s cost", name, res, got, cost)
			}
		}
	}
}

// TestRateForUnknownResolutionTakesDearest pins the anti-underquote rule: an
// unrecognised or absent resolution must never be billed at the cheap tier.
func TestRateForUnknownResolutionTakesDearest(t *testing.T) {
	m := AppModel{
		PriceUSD:          0.45,
		PerSecondPricing:  true,
		ResolutionPricing: map[string]float64{"480p": 0.20, "720p": 0.45, "1080p": 1.01, "4k": 2.31},
	}
	for _, res := range []string{"", "8k", "banana", "  "} {
		if got := m.RateFor(res); got != 2.31 {
			t.Errorf("RateFor(%q) = %.2f, want the dearest 2.31", res, got)
		}
	}
	if got := m.RateFor("1080P"); got != 1.01 { // case-insensitive
		t.Errorf("RateFor(1080P) = %.2f, want 1.01", got)
	}
	// A model without tiers is untouched.
	flat := AppModel{PriceUSD: 0.09, PerSecondPricing: true}
	if got := flat.RateFor("1080p"); got != 0.09 {
		t.Errorf("untiered model RateFor = %.2f, want its flat 0.09", got)
	}
}

// TestSeedance720pUnchanged pins that adding tiers did not reprice the default
// anyone is actually using - every seedance model defaults to 720p.
func TestSeedance720pUnchanged(t *testing.T) {
	for name, want := range map[string]float64{
		"seedance-2.0-text": 0.45, "seedance-2.0-image": 0.45,
		"seedance-2.0-fast-image": 0.40, "seedance-2.0-reference": 0.80,
	} {
		meta := modelMeta[name]
		m := AppModel{PriceUSD: meta.PriceUSD, ResolutionPricing: meta.ResolutionPricing}
		if got := m.RateFor("720p"); got != want {
			t.Errorf("%s at 720p = %.2f, want the unchanged %.2f", name, got, want)
		}
	}
}
