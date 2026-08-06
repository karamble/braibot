// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- seedance-2.0-reference ---

type seedanceReferenceModel struct{}

func (m *seedanceReferenceModel) Define() Model {
	defaultAudio := true
	return Model{
		Name:        "seedance-2.0-reference",
		Description: "ByteDance Seedance 2.0 Reference-to-Video - Generate video from text plus reference images, videos, and audio",
		Type:        "multi2video",
		Endpoint:    "https://queue.fal.run/bytedance/seedance-2.0/reference-to-video",
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

// --- minimax/h3 reference-to-video ---

type minimaxH3ReferenceModel struct{}

func (m *minimaxH3ReferenceModel) Define() Model {
	return Model{
		Name:        "minimax/h3-reference",
		Description: "MiniMax H3 reference-to-video - generate video from a prompt plus reference images, videos and audio.",
		Type:        "multi2video",
		Endpoint:    "/minimax/h3/reference-to-video",
		Options: &MinimaxH3ReferenceOptions{
			Duration:    5,
			Resolution:  "2K",
			AspectRatio: "adaptive",
		},
	}
}

func init() {
	registerModel(&minimaxH3ReferenceModel{})
}
