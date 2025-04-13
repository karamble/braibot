// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package assetserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

// Client represents a client for the braibot-assetserver
type Client struct {
	serverURL string
	apiKey    string
	client    *http.Client
}

// UploadResponse represents the response from the assetserver
type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
}

// NewClient creates a new assetserver client
func NewClient(serverURL, apiKey string) *Client {
	return &Client{
		serverURL: serverURL,
		apiKey:    apiKey,
		client:    &http.Client{},
	}
}

// UploadFile uploads a file to the assetserver and returns the download URL
func (c *Client) UploadFile(filename string, fileData []byte) (*UploadResponse, error) {
	// Create a buffer to store the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file writer for the file
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %v", err)
	}

	// Write the file data
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("failed to write file data: %v", err)
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", c.serverURL+"/upload", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse the response
	var uploadResp UploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Check if upload was successful
	if !uploadResp.Success {
		return &uploadResp, fmt.Errorf("upload failed: %s", uploadResp.Message)
	}

	return &uploadResp, nil
}
