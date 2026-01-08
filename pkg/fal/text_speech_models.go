// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- minimax-tts/text-to-speech ---

type minimaxTTSModel struct{}

func (m *minimaxTTSModel) Define() Model {
	// Define default options
	defaultOpts := &MinimaxTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()
	// Extract potentially nil defaults safely
	var defaultSpeed *float64
	if v, ok := defaults["speed"].(*float64); ok {
		defaultSpeed = v
	}
	var defaultVol *float64
	if v, ok := defaults["vol"].(*float64); ok {
		defaultVol = v
	}

	return Model{
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters",
		PriceUSD:    0.10,
		Type:        "text2speech",
		HelpDoc:     "Usage: !text2speech [text] --voice_id [voice_id] [--option value]...\nExample: !text2speech Hello world --voice_id Wise_Woman --speed 0.8 --format flac\n\nParameters:\n• text: Text to convert to speech (required)\n• --voice_id: Voice ID to use (defaults to Wise_Woman if not specified). See list below.\n• --speed: Speech speed (0.5-2.0, default: 1.0)\n• --vol: Volume (0-10, default: 1.0)\n• --pitch: Voice pitch (-12 to 12, optional)\n• --emotion: happy, sad, angry, fearful, disgusted, surprised, neutral (optional)\n• --sample_rate: 8000, 16000, 22050, 24000, 32000, 44100 (default: 32000)\n• --bitrate: 32000, 64000, 128000, 256000 (default: 128000)\n• --format: mp3, pcm, flac (default: mp3)\n• --channel: 1 (mono), 2 (stereo) (default: 1)\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl",
		Options: &MinimaxTTSOptions{
			Speed:      defaultSpeed,
			Vol:        defaultVol,
			SampleRate: defaults["sample_rate"].(string),
			Bitrate:    defaults["bitrate"].(string),
			Format:     defaults["format"].(string),
			Channel:    defaults["channel"].(string),
			// Pitch and Emotion default to nil/""
		},
	}
}

// --- chatterbox-tts ---

type chatterboxTTSModel struct{}

func (m *chatterboxTTSModel) Define() Model {
	defaultOpts := &ChatterboxTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "chatterbox-tts",
		Description: "Chatterbox TTS Turbo - Fast, natural text-to-speech",
		PriceUSD:    0.05, // Per 1000 characters
		Type:        "text2speech",
		HelpDoc:     "Usage: !text2speech [text] [options]\n\n💰 **Price: $0.05 per 1000 characters\n\nParameters:\n• text: Text to convert to speech (required)\n• --audio_prompt_url: Reference audio URL for voice cloning (optional)\n• --exaggeration: Expression intensity 0-1 (default: 0.5)\n• --cfg_weight: Adherence to prompt 0-1 (default: 0.5)",
		Options: &ChatterboxTTSOptions{
			Exaggeration: defaults["exaggeration"].(float64),
			CFGWeight:    defaults["cfg_weight"].(float64),
		},
	}
}

// --- elevenlabs-dialog ---

type elevenlabsDialogModel struct{}

func (m *elevenlabsDialogModel) Define() Model {
	defaultOpts := &ElevenLabsDialogOptions{}
	defaults := defaultOpts.GetDefaultValues()

	return Model{
		Name:        "elevenlabs-dialog",
		Description: "ElevenLabs Text-to-Dialogue V3 - Multi-speaker dialogue generation",
		PriceUSD:    0.30, // Per 1000 characters
		Type:        "text2speech",
		HelpDoc:     "Usage: !text2speech [text] [options]\n\n💰 **Price: $0.30 per 1000 characters\n\nParameters:\n• text: Dialogue text with speaker labels (required)\n• --voice_id: Voice ID (default: Rachel)\n• --output_format: Audio format (default: mp3_22050_32)\n• --stability: Voice stability 0-1 (default: 0.5)\n• --similarity_boost: Voice similarity 0-1 (default: 0.75)",
		Options: &ElevenLabsDialogOptions{
			VoiceID:         defaults["voice_id"].(string),
			OutputFormat:    defaults["output_format"].(string),
			Stability:       defaults["stability"].(float64),
			SimilarityBoost: defaults["similarity_boost"].(float64),
		},
	}
}

// --- elevenlabs-tts-turbo ---

type elevenlabsTTSTurboModel struct{}

func (m *elevenlabsTTSTurboModel) Define() Model {
	defaultOpts := &ElevenLabsTTSOptions{}
	defaults := defaultOpts.GetDefaultValues()

	var defaultStability *float64
	if v, ok := defaults["stability"].(*float64); ok {
		defaultStability = v
	}
	var defaultSimilarity *float64
	if v, ok := defaults["similarity_boost"].(*float64); ok {
		defaultSimilarity = v
	}
	var defaultStyle *float64
	if v, ok := defaults["style"].(*float64); ok {
		defaultStyle = v
	}
	var defaultSpeed *float64
	if v, ok := defaults["speed"].(*float64); ok {
		defaultSpeed = v
	}
	var defaultTimestamps *bool
	if v, ok := defaults["timestamps"].(*bool); ok {
		defaultTimestamps = v
	}

	return Model{
		Name:        "elevenlabs/tts/turbo-v2.5",
		Description: "ElevenLabs TTS Turbo v2.5 - High-quality, low-latency text-to-speech",
		PriceUSD:    0.05, // Per 1000 characters
		Type:        "text2speech",
		HelpDoc: `Usage: !text2speech [text] [options]

💰 **Price: $0.05 per 1000 characters

Parameters:
• text: Text to convert to speech (required, max 5000 chars)
• --voice: Voice name (default: Rachel)
• --stability: Voice stability 0-1 (default: 0.5)
• --similarity_boost: Voice similarity 0-1 (default: 0.75)
• --style: Style exaggeration 0-1 (default: 0.0)
• --speed: Speech speed 0.25-4.0 (default: 1.0)
• --language_code: Language code (optional)

Available Voices:
• Aria, Roger, Sarah, Laura, Charlie, George, Callum
• River, Liam, Charlotte, Alice, Matilda, Will, Jessica
• Eric, Chris, Brian, Daniel, Lily, Bill`,
		Options: &ElevenLabsTTSOptions{
			Voice:           defaults["voice"].(string),
			Stability:       defaultStability,
			SimilarityBoost: defaultSimilarity,
			Style:           defaultStyle,
			Speed:           defaultSpeed,
			Timestamps:      defaultTimestamps,
		},
	}
}

func init() {
	registerModel(&minimaxTTSModel{})
	registerModel(&chatterboxTTSModel{})
	registerModel(&elevenlabsDialogModel{})
	registerModel(&elevenlabsTTSTurboModel{})
}
