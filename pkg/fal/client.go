// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

// FinalResponseDecoder defines the function signature for decoding the final successful response
// from a Fal.ai async workflow.
type FinalResponseDecoder func(data []byte) (interface{}, error)

// executeAsyncWorkflow handles the common Fal.ai async task flow:
// 1. POST to initiate the task.
// 2. Poll the status URL until completion.
// 3. GET the final result from the response URL.
// 4. Decode the final result using the provided decoder.
func (c *Client) executeAsyncWorkflow(ctx context.Context, path string, reqBody interface{}, progress ProgressCallback, decodeFinalResponse FinalResponseDecoder) (interface{}, error) {
	// 1. Make initial POST request
	initialResp, err := c.makeRequest(ctx, "POST", path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make initial request: %w", err)
	}
	defer initialResp.Body.Close()

	if initialResp.StatusCode < 200 || initialResp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(initialResp.Body)
		return nil, fmt.Errorf("initial request failed with status %d: %s", initialResp.StatusCode, string(bodyBytes))
	}

	// 2. Parse initial QueueResponse
	var queueResp QueueResponse
	if err := json.NewDecoder(initialResp.Body).Decode(&queueResp); err != nil {
		// Attempt to read body for better error message if decode fails
		initialResp.Body.Close() // Close previous reader
		bodyBytes, readErr := io.ReadAll(initialResp.Body)
		if readErr == nil {
			return nil, fmt.Errorf("failed to decode initial queue response: %w. Body: %s", err, string(bodyBytes))
		}
		return nil, fmt.Errorf("failed to decode initial queue response: %w", err)
	}

	if queueResp.ResponseURL == "" {
		return nil, fmt.Errorf("initial queue response did not contain a response URL")
	}

	// 3. Notify initial queue position
	c.notifyQueuePosition(ctx, queueResp, progress)

	// 4. Poll queue status
	finalQueueStatus, err := c.pollQueueStatus(ctx, queueResp, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to poll queue status: %w", err)
	}

	// 5. Get final result
	finalRespRaw, err := c.makeRequest(ctx, "GET", finalQueueStatus.ResponseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get final result: %w", err)
	}
	defer finalRespRaw.Body.Close()

	finalBytes, err := io.ReadAll(finalRespRaw.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read final result body: %w", err)
	}

	if finalRespRaw.StatusCode < 200 || finalRespRaw.StatusCode >= 300 {
		return nil, fmt.Errorf("final result request failed with status %d: %s", finalRespRaw.StatusCode, string(finalBytes))
	}

	if c.debug {
		fmt.Printf("DEBUG - Final response body: %s\n", string(finalBytes))
	}

	// 6. Decode final response using the provided decoder function
	finalData, err := decodeFinalResponse(finalBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode final response: %w", err)
	}

	return finalData, nil
}
