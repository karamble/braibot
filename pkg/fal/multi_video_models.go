// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- seedance-2.0-reference ---

type seedanceReferenceModel struct{}

func (m *seedanceReferenceModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:             "seedance-2.0-reference",
		Description:      "ByteDance Seedance 2.0 Reference-to-Video - Generate video from text plus reference images, videos, and audio",
		PriceUSD:         0.80, // $0.80 per second (flat rate; covers worst-case 15s reference video input at 720p which costs fal ~$0.7258/s effective)
		Type:             "multi2video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !multi2video [prompt] [options]\n\n💰 **Price: $0.80 per second**\nExample: A 5-second video will cost $4.00.\nTotal cost = price per second × duration.\n\nParameters:\n• prompt: Text description of the desired video (required)\n• --image1..--image9: Reference image URLs (up to 9, JPEG/PNG/WebP, max 30MB each)\n• --video1..--video3: Reference video URLs (up to 3, MP4/MOV, 2-15s combined duration, <50MB total, 480p-720p)\n• --audio1..--audio3: Reference audio URLs (up to 3, MP3/WAV, ≤15s combined, max 15MB each)\n• --duration: Output video duration in seconds (4-15, default: 5)\n• --aspect: Aspect ratio (auto, 21:9, 16:9, 4:3, 1:1, 3:4, 9:16). Default: auto\n• --resolution: Output video resolution (480p, 720p). Default: 720p\n• --audio: Enable generated audio output (default: true)\n• --seed: Seed for reproducibility (optional)\n\nConstraints:\n• At least one reference input (image, video, or audio) is required\n• Total reference files must not exceed 12\n• Reference audio requires at least one reference image or video",
		Options: &SeedanceReferenceOptions{
			Duration:      "5",
			AspectRatio:   "auto",
			Resolution:    "720p",
			GenerateAudio: &defaultAudio,
		},
	}
}

func init() {
	registerModel(&seedanceReferenceModel{})
}
