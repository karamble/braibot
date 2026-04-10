// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

// --- ghiblify ---

type ghiblifyModel struct{}

func (m *ghiblifyModel) Define() Model {
	return Model{
		Name:        "ghiblify",
		Description: "Transforms images into Studio Ghibli style artwork",
		Type:        "image2image",
		Endpoint:    "/ghiblify",
	}
}

// --- cartoonify ---

type cartoonifyModel struct{}

func (m *cartoonifyModel) Define() Model {
	return Model{
		Name:        "cartoonify",
		Description: "Transforms images into Pixar like 3d cartoon-style artwork",
		Type:        "image2image",
		Endpoint:    "/cartoonify",
		Options:     &CartoonifyOptions{},
	}
}

// --- star-vector ---

type starVectorModel struct{}

func (m *starVectorModel) Define() Model {
	return Model{
		Name:        "star-vector",
		Description: "Convert images to SVG using AI vectorization",
		Type:        "image2image",
		Endpoint:    "/fal-ai/star-vector",
		Options:     &StarVectorOptions{},
	}
}

func init() {
	registerModel(&ghiblifyModel{})
	registerModel(&cartoonifyModel{})
	registerModel(&starVectorModel{})
}
