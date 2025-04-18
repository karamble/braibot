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
		HelpDoc:     "Usage: !text2speech [voice_id] [text] [--option value]...\nExample: !text2speech Wise_Woman Hello --speed 0.8 --format flac\n\nParameters:\n• voice_id: Optional voice ID (defaults to Wise_Woman). See list below.\n• text: Text to convert to speech (required)\n• --speed: Speech speed (0.5-2.0, default: 1.0)\n• --vol: Volume (0-10, default: 1.0)\n• --pitch: Voice pitch (-12 to 12, optional)\n• --emotion: happy, sad, angry, fearful, disgusted, surprised, neutral (optional)\n• --sample_rate: 8000, 16000, 22050, 24000, 32000, 44100 (default: 32000)\n• --bitrate: 32000, 64000, 128000, 256000 (default: 128000)\n• --format: mp3, pcm, flac (default: mp3)\n• --channel: 1 (mono), 2 (stereo) (default: 1)\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl",
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

func init() {
	registerModel(&minimaxTTSModel{})
}
