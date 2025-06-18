// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"fmt"
)

// --- kling-video-text ---

type klingVideoTextModel struct{}

func (m *klingVideoTextModel) Define() Model {
	return Model{
		Name:        "kling-video-text",
		Description: "Generate videos from text using Kling 2.0 Master.",
		PriceUSD:    0.4,
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.40 per video.",
		Options: &KlingVideoOptions{
			Duration:       "5",
			AspectRatio:    "16:9",
			NegativePrompt: "blur, distort, and low quality",
			CFGScale:       0.5,
		},
		PerSecondPricing: true,
	}
}

func init() {
	registerModel(&klingVideoTextModel{})
}

// --- minimax-video-01-director ---

type minimaxDirectorModel struct{}

func (m *minimaxDirectorModel) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01-director",
		Description: "Generate video clips with camera movement instructions.",
		PriceUSD:    0.8, // Per second
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.80 per video.",
		Options: &MinimaxDirectorOptions{
			PromptOptimizer: &defaultOptimizer,
		},
		PerSecondPricing: false,
	}
}

func init() {
	registerModel(&minimaxDirectorModel{})
}

// --- minimax/video-01 ---

type minimaxVideo01Model struct{}

func (m *minimaxVideo01Model) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:        "minimax/video-01",
		Description: "Native high-resolution, high-frame-rate video generation model.",
		PriceUSD:    0.8, // Per second
		Type:        "text2video",
		HelpDoc:     "Usage: !text2video [prompt] [options]\n\n💰 **Price: $0.80 per video",
		Options: &MinimaxVideo01Options{
			PromptOptimizer: &defaultOptimizer,
		},
		PerSecondPricing: false,
	}
}

func init() {
	registerModel(&minimaxVideo01Model{})
}

// MiniMax Hailuo-02 Text To Video Model

type minimaxHailuo02Model struct{}

type MinimaxHailuo02Options struct {
	Duration        string `json:"duration,omitempty"` // "6" or "10"
	PromptOptimizer *bool  `json:"prompt_optimizer,omitempty"`
}

type VideoOptions = MinimaxHailuo02Options

func (o *MinimaxHailuo02Options) GetDefaultValues() map[string]interface{} {
	defaultOptimizer := true
	return map[string]interface{}{
		"duration":         "6",
		"prompt_optimizer": &defaultOptimizer,
	}
}

func (o *MinimaxHailuo02Options) Validate() error {
	if o.Duration != "" && o.Duration != "6" && o.Duration != "10" {
		return fmt.Errorf("invalid duration: %s (must be 6 or 10)", o.Duration)
	}
	return nil
}

func (m *minimaxHailuo02Model) Define() Model {
	defaultOptimizer := true
	return Model{
		Name:             "minimax/hailuo-02",
		Description:      "MiniMax Hailuo-02 Text To Video. Per-second pricing.",
		PriceUSD:         0.09,
		Type:             "text2video",
		HelpDoc:          "Usage: !text2video [prompt] [--duration 6|10] [--prompt_optimizer true|false]\n\n💰 **Price: $0.10 per video second**\nExample: A 10-second video will cost $1.00.\nTotal cost = price per second × duration.",
		Options:          &MinimaxHailuo02Options{Duration: "6", PromptOptimizer: &defaultOptimizer},
		PerSecondPricing: true,
	}
}

func init() {
	registerModel(&minimaxHailuo02Model{})
}
