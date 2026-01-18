// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- elevenlabs-voice-changer ---

type elevenlabsVoiceChangerModel struct{}

func (m *elevenlabsVoiceChangerModel) Define() Model {
	defaultOpts := &ElevenLabsVoiceChangerOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:             "elevenlabs-voice-changer",
		Description:      "ElevenLabs Voice Changer - Transform audio with different voices",
		PriceUSD:         0.50, // Placeholder: actual is $0.30/min
		Type:             "audio2audio",
		PerSecondPricing: true,
		HelpDoc: `Usage: !audio2audio [audio_url] [options]

Price: $0.30 per minute

Parameters:
- audio_url: URL of audio to transform (required)
- --voice: Voice name (default: Rachel)
- --remove_background_noise: Remove background noise (optional)
- --seed: Random seed for reproducibility (optional)
- --output_format: Output format (default: mp3_44100_128)

Available Voices:
- Aria, Roger, Sarah, Laura, Charlie, George, Callum
- River, Liam, Charlotte, Alice, Matilda, Will, Jessica
- Eric, Chris, Brian, Daniel, Lily, Bill, Rachel`,
		Options: &ElevenLabsVoiceChangerOptions{
			Voice:        defaults["voice"].(string),
			OutputFormat: defaults["output_format"].(string),
		},
	}
}

func init() {
	registerModel(&elevenlabsVoiceChangerModel{})
}
