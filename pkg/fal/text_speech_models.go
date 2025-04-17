// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- minimax-tts/text-to-speech ---

type minimaxTTSModel struct{}

func (m *minimaxTTSModel) Define() Model {
	return Model{
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters",
		PriceUSD:    0.10,
		Type:        "text2speech",
		HelpDoc:     "Usage: !text2speech [voice_id] [text]\nExample: !text2speech Wise_Woman Hello, how are you today?\n\nParameters:\n• voice_id: Optional voice ID (defaults to Wise_Woman)\n• text: Text to convert to speech\n\nAvailable Voices:\n• Wise_Woman, Friendly_Person, Inspirational_girl\n• Deep_Voice_Man, Calm_Woman, Casual_Guy\n• Lively_Girl, Patient_Man, Young_Knight\n• Determined_Man, Lovely_Girl, Decent_Boy\n• Imposing_Manner, Elegant_Man, Abbess\n• Sweet_Girl_2, Exuberant_Girl",
		// Note: No specific options struct defined for this model in the original code.
		Options: nil,
	}
}

func init() {
	registerModel(&minimaxTTSModel{})
}
