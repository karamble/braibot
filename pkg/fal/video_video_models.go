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
		Type:        "video2video",
		Endpoint:    "/topaz/upscale/video",
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
		Type:        "video2video",
		Endpoint:    "/sync-lipsync/v2",
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
		Name:        "kling-video-v26-motion-control",
		Description: "Kling Video v2.6 Motion Control - Generate videos with character motion from reference video",
		Type:        "video2video",
		Endpoint:    "/kling-video/v2.6/standard/motion-control",
		Options: &KlingVideoV26MotionControlOptions{
			CharacterOrientation: defaults["character_orientation"].(string),
			KeepOriginalSound:    defaultKeepSound,
		},
	}
}

// --- kling-video-o3-edit ---

type klingVideoO3EditModel struct{}

func (m *klingVideoO3EditModel) Define() Model {
	defaultKeepAudio := true
	return Model{
		Name:        "kling-video-o3-edit",
		Description: "Kling O3 Standard Video Edit - Edit videos with multi-scene consistency",
		Type:        "video2video",
		Endpoint:    "/kling-video/o3/standard/video-to-video/edit",
		Options: &KlingVideoO3EditOptions{
			KeepAudio: &defaultKeepAudio,
		},
	}
}

// --- kling-video-o3-pro-edit ---

type klingVideoO3ProEditModel struct{}

func (m *klingVideoO3ProEditModel) Define() Model {
	defaultKeepAudio := true
	return Model{
		Name:        "kling-video-o3-pro-edit",
		Description: "Kling O3 Pro Video Edit - Premium video editing with multi-scene consistency",
		Type:        "video2video",
		Endpoint:    "/kling-video/o3/pro/video-to-video/edit",
		Options: &KlingVideoO3EditOptions{
			KeepAudio: &defaultKeepAudio,
		},
	}
}

func init() {
	registerModel(&topazUpscaleVideoModel{})
	registerModel(&syncLipsyncV2Model{})
	registerModel(&klingVideoV26MotionControlModel{})
	registerModel(&klingVideoO3EditModel{})
	registerModel(&klingVideoO3ProEditModel{})
}
