// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import "testing"

// TestMinimaxH3ModelsRegister pins that all three H3 variants are reachable
// under the type the request dispatcher looks them up with - a mismatch there
// fails only at call time, with "model not found".
func TestMinimaxH3ModelsRegister(t *testing.T) {
	for _, tc := range []struct{ name, modelType, endpoint string }{
		{"minimax/h3-text", "text2video", "/minimax/h3/text-to-video"},
		{"minimax/h3-image", "image2video", "/minimax/h3/image-to-video"},
		{"minimax/h3-reference", "multi2video", "/minimax/h3/reference-to-video"},
	} {
		m, ok := GetModel(tc.name, tc.modelType)
		if !ok {
			t.Errorf("%s not registered as %s", tc.name, tc.modelType)
			continue
		}
		if m.Endpoint != tc.endpoint {
			t.Errorf("%s endpoint = %q, want %q", tc.name, m.Endpoint, tc.endpoint)
		}
		if m.Options == nil {
			t.Errorf("%s has no default options", tc.name)
		}
	}
}

// TestMinimaxH3Defaults checks the defaults match fal's documented ones. The
// resolution default matters for money: 2K bills at $0.26/s against 768P's
// $0.16/s, and the registry price assumes the dearer tier.
func TestMinimaxH3Defaults(t *testing.T) {
	text := (&MinimaxH3Options{}).GetDefaultValues()
	if text["resolution"] != "2K" || text["duration"] != 5 || text["aspect_ratio"] != "16:9" {
		t.Errorf("text-to-video defaults = %v", text)
	}
	img := (&MinimaxH3ImageOptions{}).GetDefaultValues()
	if img["resolution"] != "2K" || img["duration"] != 5 {
		t.Errorf("image-to-video defaults = %v", img)
	}
	ref := (&MinimaxH3ReferenceOptions{}).GetDefaultValues()
	if ref["resolution"] != "2K" || ref["aspect_ratio"] != "adaptive" {
		t.Errorf("reference-to-video defaults = %v", ref)
	}
}

// TestMinimaxH3Validate covers the enums. Resolution selects the billing tier,
// so a typo must be rejected rather than silently priced at the wrong rate.
func TestMinimaxH3Validate(t *testing.T) {
	for _, res := range []string{"768P", "2K", "768p", "2k", ""} {
		if err := (&MinimaxH3Options{Resolution: res}).Validate(); err != nil {
			t.Errorf("resolution %q should be valid: %v", res, err)
		}
	}
	for _, res := range []string{"1080p", "4K", "720P", "banana"} {
		if err := (&MinimaxH3Options{Resolution: res}).Validate(); err == nil {
			t.Errorf("resolution %q should be rejected", res)
		}
	}

	if err := (&MinimaxH3Options{Duration: 11}).Validate(); err == nil {
		t.Error("duration 11 should be rejected")
	}
	if err := (&MinimaxH3Options{Duration: 5}).Validate(); err != nil {
		t.Errorf("duration 5 should be valid: %v", err)
	}

	// "adaptive" is a reference-to-video-only aspect ratio.
	if err := (&MinimaxH3ReferenceOptions{AspectRatio: "adaptive"}).Validate(); err != nil {
		t.Errorf("adaptive should be valid on reference-to-video: %v", err)
	}
	if err := (&MinimaxH3Options{AspectRatio: "adaptive"}).Validate(); err == nil {
		t.Error("adaptive should be rejected on text-to-video")
	}
}
