// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- fast-sdxl ---

type fastSDXLModel struct{}

func (m *fastSDXLModel) Define() Model {
	return Model{
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly",
		PriceUSD:    0.02,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- hidream-i1-full ---

type hidreamI1FullModel struct{}

func (m *hidreamI1FullModel) Define() Model {
	return Model{
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images",
		PriceUSD:    0.10,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- hidream-i1-dev ---

type hidreamI1DevModel struct{}

func (m *hidreamI1DevModel) Define() Model {
	return Model{
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model",
		PriceUSD:    0.06,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- hidream-i1-fast ---

type hidreamI1FastModel struct{}

func (m *hidreamI1FastModel) Define() Model {
	return Model{
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model",
		PriceUSD:    0.03,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- flux-pro/v1.1 ---

type fluxProV1_1Model struct{}

func (m *fluxProV1_1Model) Define() Model {
	return Model{
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation",
		PriceUSD:    0.08,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- flux-pro/v1.1-ultra ---

type fluxProV1_1UltraModel struct{}

func (m *fluxProV1_1UltraModel) Define() Model {
	return Model{
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model",
		PriceUSD:    0.12,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

// --- flux/schnell ---

type fluxSchnellModel struct{}

func (m *fluxSchnellModel) Define() Model {
	return Model{
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation",
		PriceUSD:    0.02,
		Type:        "text2image",
		HelpDoc:     "Usage: !text2image \nExample: !text2image a beautiful sunset over mountains\n\nParameters:\n• prompt: Text description of the image you want to generate",
	}
}

func init() {
	registerModel(&fastSDXLModel{})
	registerModel(&hidreamI1FullModel{})
	registerModel(&hidreamI1DevModel{})
	registerModel(&hidreamI1FastModel{})
	registerModel(&fluxProV1_1Model{})
	registerModel(&fluxProV1_1UltraModel{})
	registerModel(&fluxSchnellModel{})
}
