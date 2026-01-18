// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- topaz-upscale-video ---

type topazUpscaleVideoModel struct{}

func (m *topazUpscaleVideoModel) Define() Model {
	defaultOpts := &TopazUpscaleVideoOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "topaz-upscale-video",
		Description: "Topaz Video Upscale - Professional-grade video upscaling",
		PriceUSD:    0.50,
		Type:        "video2video",
		HelpDoc:     "Usage: !video2video [video_url] [options]\n\n💰 **Price: $0.50 per video\n\nParameters:\n• video_url: URL of the video to upscale\n• --model: Upscaling model (default: auto)\n• --output_type: Output format mp4 or mov (default: mp4)",
		Options: &TopazUpscaleVideoOptions{
			Model:      defaults["model"].(string),
			OutputType: defaults["output_type"].(string),
		},
	}
}

// --- sync-lipsync-v2 ---

type syncLipsyncV2Model struct{}

func (m *syncLipsyncV2Model) Define() Model {
	defaultOpts := &SyncLipsyncV2Options{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "sync-lipsync-v2",
		Description: "Sync Lipsync V2 - Generate realistic lipsync animations from audio",
		PriceUSD:    0.10, // Per second
		Type:        "video2video",
		PerSecondPricing: true,
		HelpDoc:     "Usage: !video2video [video_url] [audio_url] [options]\n\n💰 **Price: $0.10 per second\n\nParameters:\n• video_url: URL of the video with face\n• audio_url: URL of the audio to sync\n• --model: wav2lip or wav2lip_gan (default: wav2lip)\n• --output_type: Output format mp4 or webm (default: mp4)",
		Options: &SyncLipsyncV2Options{
			Model:      defaults["model"].(string),
			OutputType: defaults["output_type"].(string),
		},
	}
}

// --- kling-video-v26-motion-control ---

type klingVideoV26MotionControlModel struct{}

func (m *klingVideoV26MotionControlModel) Define() Model {
	defaultOpts := &KlingVideoV26MotionControlOptions{}
	defaults := defaultOpts.GetDefaultValues()
	defaultKeepSound := defaults["keep_original_sound"].(*bool)

	return Model{
		Name:             "kling-video-v26-motion-control",
		Description:      "Kling Video v2.6 Motion Control - Generate videos with character motion from reference video",
		PriceUSD:         1.50, // Placeholder: $0.10/sec, using 15s average estimate
		Type:             "video2video",
		PerSecondPricing: true,
		HelpDoc:          "Usage: !video2video [image_url] [video_url] [options]\n\n💰 **Price: $0.10 per second\n\nParameters:\n• image_url: Reference image URL (character/background source)\n• video_url: Reference video URL (motion source)\n• --prompt: Text description (optional)\n• --orientation: 'image' (max 10s) or 'video' (max 30s). Default: video\n• --keep-sound: Keep original audio (default: true)\n\nConstraints:\n• Character must occupy >5% of image with visible body",
		Options: &KlingVideoV26MotionControlOptions{
			CharacterOrientation: defaults["character_orientation"].(string),
			KeepOriginalSound:    defaultKeepSound,
		},
	}
}

func init() {
	registerModel(&topazUpscaleVideoModel{})
	registerModel(&syncLipsyncV2Model{})
	registerModel(&klingVideoV26MotionControlModel{})
}
