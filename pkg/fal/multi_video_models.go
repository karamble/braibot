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
