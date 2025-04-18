// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- ghiblify ---

type ghiblifyModel struct{}

func (m *ghiblifyModel) Define() Model {
	return Model{
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
	}
}

// --- cartoonify ---

type cartoonifyModel struct{}

func (m *cartoonifyModel) Define() Model {
	return Model{
		Name:        "cartoonify",
		Description: "Transforms images into Pixar like 3d cartoon-style artwork",
		PriceUSD:    0.02,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nParameters:\n• image_url: URL of the image to transform",
		Options:     &CartoonifyOptions{},
	}
}

// --- star-vector ---

type starVectorModel struct{}

func (m *starVectorModel) Define() Model {
	return Model{
		Name:        "star-vector",
		Description: "Convert images to SVG using AI vectorization",
		PriceUSD:    1.0,
		Type:        "image2image",
		HelpDoc:     "Usage: !image2image [image_url]\nExample: !image2image https://example.com/image.jpg\n\nTo use this model, first set it as the default model:\n!setmodel image2image star-vector\n\nParameters:\n• image_url: URL of the source image\n\nPricing:\n• Base price: $1.0 per image",
		Options:     &StarVectorOptions{},
	}
}

func init() {
	registerModel(&ghiblifyModel{})
	registerModel(&cartoonifyModel{})
	registerModel(&starVectorModel{})
}
