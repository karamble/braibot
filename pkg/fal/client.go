// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	baseURL = "https://queue.fal.run/fal-ai"
)

// Client represents a Fal.ai API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	debug      bool
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithDebug enables debug mode for the client
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new Fal.ai API client
func NewClient(apiKey string, opts ...ClientOption) *Client {
	client := &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// makeRequest makes an HTTP request to the Fal.ai API
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
	}

	// Check if path is a full URL
	var fullURL string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		fullURL = path
	} else {
		fullURL = baseURL + path
	}

	if c.debug {
		fmt.Printf("DEBUG - Making request to Fal.ai API:\n")
		fmt.Printf("  URL: %s\n", fullURL)
		fmt.Printf("  Method: %s\n", method)
		if body != nil {
			fmt.Printf("  Request Body: %s\n", string(reqBody))
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}

	if c.debug {
		fmt.Printf("DEBUG - Response from Fal.ai API:\n")
		fmt.Printf("  Status Code: %d\n", resp.StatusCode)
		fmt.Printf("  Status: %s\n", resp.Status)
	}

	return resp, nil
}
