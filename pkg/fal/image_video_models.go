// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- veo2 ---

type veo2Model struct{}

func (m *veo2Model) Define() Model {
	return Model{
		Name:        "veo2",
		Description: "Creates videos from images with realistic motion using Google's Veo 2 model. Base price: $2.50 for 5 seconds, $0.50 per additional second",
		PriceUSD:    3.50,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --aspect 16:9 --duration 5\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --aspect: Aspect ratio (16:9, 9:16, 1:1)\n• --duration: Video duration (5, 6, 7, 8)\n\nPricing:\n• Base price: $3.50 for 5 seconds\n• Additional cost: $0.50 per second beyond 5 seconds",
		Options: &Veo2Options{
			AspectRatio: "16:9",
			Duration:    "5",
		},
	}
}

// --- kling-video-image ---

type klingVideoImageModel struct{}

func (m *klingVideoImageModel) Define() Model {
	return Model{
		Name:        "kling-video-image",
		Description: "Convert images to video using Kling 2.0 Master. Base price: $2.0 for 5 seconds, $0.4 per additional second",
		PriceUSD:    2.0,
		Type:        "image2video",
		HelpDoc:     "Usage: !image2video [image_url] [prompt] [options]\nExample: !image2video https://example.com/image.jpg a beautiful animation --duration 10 --aspect 16:9\n\nParameters:\n• image_url: URL of the source image\n• prompt: Description of the desired video animation\n• --duration: Video duration in seconds (default: 5, min: 5)\n• --aspect: Aspect ratio (default: 16:9)\n• --negative-prompt: Text describing what to avoid (default: blur, distort, and low quality)\n• --cfg-scale: Configuration scale (default: 0.5)\n\nPricing:\n• Base price: $2.0 for 5 seconds\n• Additional cost: $0.4 per second beyond 5 seconds",
		Options: &KlingVideoOptions{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
		},
	}
}

func init() {
	registerModel(&veo2Model{})
	registerModel(&klingVideoImageModel{})
}
