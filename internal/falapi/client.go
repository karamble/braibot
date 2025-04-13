package falapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	baseURL = "https://queue.fal.run/fal-ai"
)

type Text2ImageResponse struct {
	Images []struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
	} `json:"images"`
}

type QueueResponse struct {
	Status      string `json:"status"`
	RequestID   string `json:"request_id"`
	ResponseURL string `json:"response_url"`
	StatusURL   string `json:"status_url"`
	CancelURL   string `json:"cancel_url"`
	Logs        []struct {
		Message   string `json:"message"`
		Level     string `json:"level"`
		Source    string `json:"source"`
		Timestamp string `json:"timestamp"`
	} `json:"logs"`
	QueuePosition int `json:"queue_position"`
}

// Client represents a Fal.ai API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	debug      bool
}

// NewClient creates a new Fal.ai API client
func NewClient(apiKey string, debug bool) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		debug: debug,
	}
}

// GenerateImage generates an image from a text prompt
func (c *Client) GenerateImage(ctx context.Context, prompt string, modelName string, bot interface{}, userNick string) (*ImageResponse, error) {
	// Create request body
	reqBody := map[string]interface{}{
		"prompt": prompt,
	}

	// Make initial request to queue
	resp, err := c.makeRequest(ctx, "POST", "/"+modelName, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the initial queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}
	resp.Body.Close()

	// If we have a queue position, notify the user
	if queueResp.QueuePosition >= 0 {
		// Use reflection to call SendPM on the bot
		botValue := reflect.ValueOf(bot)
		sendPMMethod := botValue.MethodByName("SendPM")
		if sendPMMethod.IsValid() {
			message := "Your image generation request is at the front of the queue."
			if queueResp.QueuePosition > 0 {
				message = fmt.Sprintf("Your image generation request is in queue. Position: %d", queueResp.QueuePosition)
			}
			sendPMMethod.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(userNick),
				reflect.ValueOf(message),
			})
		}
	}

	// Create a ticker for polling
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Poll status until completion
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Get status using status_url
			statusURL := queueResp.StatusURL + "?logs=1"
			req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create status request: %v", err)
			}
			req.Header.Set("Authorization", "Key "+c.apiKey)

			statusResp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %v", err)
			}
			defer statusResp.Body.Close()

			// Read status response body
			statusBytes, err := io.ReadAll(statusResp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read status response: %v", err)
			}

			if c.debug {
				fmt.Printf("Raw status response: %s\n", string(statusBytes))
			}

			var status struct {
				Status      string `json:"status"`
				RequestID   string `json:"request_id"`
				ResponseURL string `json:"response_url"`
				StatusURL   string `json:"status_url"`
				CancelURL   string `json:"cancel_url"`
				Logs        []struct {
					Message   string `json:"message"`
					Level     string `json:"level"`
					Source    string `json:"source"`
					Timestamp string `json:"timestamp"`
				} `json:"logs"`
				QueuePosition int `json:"queue_position"`
			}

			if err := json.Unmarshal(statusBytes, &status); err != nil {
				return nil, fmt.Errorf("failed to parse status response: %v", err)
			}

			if c.debug {
				fmt.Printf("Status Response: %+v\n", status)
			}

			switch status.Status {
			case "IN_QUEUE":
				if status.QueuePosition > 0 {
					// Send queue position update to user
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						ctxValue := reflect.ValueOf(ctx)
						userNickValue := reflect.ValueOf(userNick)
						messageValue := reflect.ValueOf(fmt.Sprintf("Your image generation request is in queue, position: %d", status.QueuePosition))

						args := []reflect.Value{ctxValue, userNickValue, messageValue}
						sendPMMethod.Call(args)
					}
				}
			case "IN_PROGRESS":
				// Send progress update to user if we have logs
				if len(status.Logs) > 0 {
					// Get the latest log entry
					latestLog := status.Logs[len(status.Logs)-1]

					// Use reflection to call SendPM on the bot
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						sendPMMethod.Call([]reflect.Value{
							reflect.ValueOf(ctx),
							reflect.ValueOf(userNick),
							reflect.ValueOf(fmt.Sprintf("Progress: %s", latestLog.Message)),
						})
					}
				}

				if len(status.Logs) > 0 {
					for _, log := range status.Logs {
						fmt.Printf("Progress: %s\n", log.Message)
					}
				}
			case "COMPLETED":
				// Get final response using the base request URL (without /status)
				responseURL := strings.TrimSuffix(queueResp.ResponseURL, "/status")
				req, err := http.NewRequestWithContext(ctx, "GET", responseURL, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create final response request: %v", err)
				}
				req.Header.Set("Authorization", "Key "+c.apiKey)

				finalResp, err := c.httpClient.Do(req)
				if err != nil {
					return nil, fmt.Errorf("failed to get final response: %v", err)
				}
				defer finalResp.Body.Close()

				// Read final response body
				finalBytes, err := io.ReadAll(finalResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read final response: %v", err)
				}

				if c.debug {
					fmt.Printf("Raw final response: %s\n", string(finalBytes))
				}

				var finalResponse struct {
					Images []struct {
						URL         string `json:"url"`
						Width       int    `json:"width"`
						Height      int    `json:"height"`
						ContentType string `json:"content_type"`
					} `json:"images"`
					Timings struct {
						Inference float64 `json:"inference"`
					} `json:"timings"`
					Seed            json.Number `json:"seed"`
					HasNSFWConcepts []bool      `json:"has_nsfw_concepts"`
					Prompt          string      `json:"prompt"`
				}

				if err := json.Unmarshal(finalBytes, &finalResponse); err != nil {
					return nil, fmt.Errorf("failed to parse final response: %v", err)
				}

				if c.debug {
					fmt.Printf("Final Response: %+v\n", finalResponse)
				}

				if len(finalResponse.Images) == 0 {
					return nil, fmt.Errorf("no images in response")
				}

				// Download the image from the URL
				imgResp, err := http.Get(finalResponse.Images[0].URL)
				if err != nil {
					return nil, fmt.Errorf("failed to download image: %v", err)
				}
				defer imgResp.Body.Close()

				// Read the image data
				imgData, err := io.ReadAll(imgResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read image data: %v", err)
				}

				// Encode the image as base64
				base64Image := base64.StdEncoding.EncodeToString(imgData)

				// Create the PM message format
				message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--",
					url.QueryEscape(prompt),
					finalResponse.Images[0].ContentType,
					base64Image)

				if c.debug {
					fmt.Printf("PM Message: %s\n", message)
				}

				// Create an ImageResponse to return
				imageResp := &ImageResponse{
					Images: []struct {
						URL         string `json:"url"`
						Width       int    `json:"width"`
						Height      int    `json:"height"`
						ContentType string `json:"content_type"`
					}{
						{
							URL:         finalResponse.Images[0].URL,
							Width:       finalResponse.Images[0].Width,
							Height:      finalResponse.Images[0].Height,
							ContentType: finalResponse.Images[0].ContentType,
						},
					},
					Timings: struct {
						Inference float64 `json:"inference"`
					}{
						Inference: finalResponse.Timings.Inference,
					},
					Seed:            finalResponse.Seed,
					HasNSFWConcepts: finalResponse.HasNSFWConcepts,
					Prompt:          finalResponse.Prompt,
				}

				return imageResp, nil
			case "FAILED":
				// Send failure message to user
				botValue := reflect.ValueOf(bot)
				sendPMMethod := botValue.MethodByName("SendPM")
				if sendPMMethod.IsValid() {
					ctxValue := reflect.ValueOf(ctx)
					userNickValue := reflect.ValueOf(userNick)
					messageValue := reflect.ValueOf("Your image generation request failed.")

					args := []reflect.Value{ctxValue, userNickValue, messageValue}
					sendPMMethod.Call(args)
				}
				return nil, fmt.Errorf("request failed")
			}
		}
	}
}

// GenerateSpeech generates speech from text
func (c *Client) GenerateSpeech(ctx context.Context, text string, voiceID string, bot interface{}, userNick string) (*AudioResponse, error) {
	// Create request body
	reqBody := map[string]interface{}{
		"text": text,
		"audio_setting": map[string]interface{}{
			"format":      "pcm",
			"sample_rate": 24000,
			"channel":     1,
		},
	}

	// Only include voice_id if it's a valid voice ID
	validVoices := map[string]bool{
		"Wise_Woman":         true,
		"Friendly_Person":    true,
		"Inspirational_girl": true,
		"Deep_Voice_Man":     true,
		"Calm_Woman":         true,
		"Casual_Guy":         true,
		"Lively_Girl":        true,
		"Patient_Man":        true,
		"Young_Knight":       true,
		"Determined_Man":     true,
		"Lovely_Girl":        true,
		"Decent_Boy":         true,
		"Imposing_Manner":    true,
		"Elegant_Man":        true,
		"Abbess":             true,
		"Sweet_Girl_2":       true,
		"Exuberant_Girl":     true,
	}

	// If voiceID is provided and valid, include it in the voice_setting
	if voiceID != "" && validVoices[voiceID] {
		reqBody["voice_setting"] = map[string]interface{}{
			"voice_id": voiceID,
		}
	}

	// Debug log for request body
	if c.debug {
		reqBodyJSON, _ := json.MarshalIndent(reqBody, "", "  ")
		fmt.Printf("Text-to-speech request body: %s\n", string(reqBodyJSON))
	}

	// Make initial request to queue with the correct URL format
	// Format: https://queue.fal.run/fal-ai/minimax-tts/text-to-speech
	resp, err := c.makeRequest(ctx, "POST", "/minimax-tts/text-to-speech", reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the initial queue response
	var queueResp QueueResponse
	if err := json.NewDecoder(resp.Body).Decode(&queueResp); err != nil {
		return nil, fmt.Errorf("failed to decode queue response: %v", err)
	}
	resp.Body.Close()

	// If we have a queue position, notify the user
	if queueResp.QueuePosition >= 0 {
		// Use reflection to call SendPM on the bot
		botValue := reflect.ValueOf(bot)
		sendPMMethod := botValue.MethodByName("SendPM")
		if sendPMMethod.IsValid() {
			message := "Your speech generation request is at the front of the queue."
			if queueResp.QueuePosition > 0 {
				message = fmt.Sprintf("Your speech generation request is in queue. Position: %d", queueResp.QueuePosition)
			}
			sendPMMethod.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(userNick),
				reflect.ValueOf(message),
			})
		}
	}

	// Create a ticker for polling
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Poll status until completion
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Get status using status_url
			statusURL := queueResp.StatusURL + "?logs=1"
			req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create status request: %v", err)
			}
			req.Header.Set("Authorization", "Key "+c.apiKey)

			statusResp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to get status: %v", err)
			}
			defer statusResp.Body.Close()

			// Read status response body
			statusBytes, err := io.ReadAll(statusResp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read status response: %v", err)
			}

			if c.debug {
				fmt.Printf("Raw status response: %s\n", string(statusBytes))
			}

			var status struct {
				Status      string `json:"status"`
				RequestID   string `json:"request_id"`
				ResponseURL string `json:"response_url"`
				StatusURL   string `json:"status_url"`
				CancelURL   string `json:"cancel_url"`
				Logs        []struct {
					Message   string `json:"message"`
					Level     string `json:"level"`
					Source    string `json:"source"`
					Timestamp string `json:"timestamp"`
				} `json:"logs"`
				QueuePosition int `json:"queue_position"`
			}

			if err := json.Unmarshal(statusBytes, &status); err != nil {
				return nil, fmt.Errorf("failed to parse status response: %v", err)
			}

			if c.debug {
				fmt.Printf("Status Response: %+v\n", status)
			}

			switch status.Status {
			case "IN_QUEUE":
				if status.QueuePosition > 0 {
					// Send queue position update to user
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						sendPMMethod.Call([]reflect.Value{
							reflect.ValueOf(ctx),
							reflect.ValueOf(userNick),
							reflect.ValueOf(fmt.Sprintf("Your speech generation request is in queue, position: %d", status.QueuePosition)),
						})
					}
				}
			case "IN_PROGRESS":
				// Send progress update to user if we have logs
				if len(status.Logs) > 0 {
					// Get the latest log entry
					latestLog := status.Logs[len(status.Logs)-1]

					// Use reflection to call SendPM on the bot
					botValue := reflect.ValueOf(bot)
					sendPMMethod := botValue.MethodByName("SendPM")
					if sendPMMethod.IsValid() {
						sendPMMethod.Call([]reflect.Value{
							reflect.ValueOf(ctx),
							reflect.ValueOf(userNick),
							reflect.ValueOf(fmt.Sprintf("Progress: %s", latestLog.Message)),
						})
					}
				}
			case "COMPLETED":
				// Get final response using the correct URL format for text2speech
				// Format: https://queue.fal.run/fal-ai/minimax-tts/requests/{request_id}
				responseURL := fmt.Sprintf("https://queue.fal.run/fal-ai/minimax-tts/requests/%s", status.RequestID)
				req, err := http.NewRequestWithContext(ctx, "GET", responseURL, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create final response request: %v", err)
				}
				req.Header.Set("Authorization", "Key "+c.apiKey)

				finalResp, err := c.httpClient.Do(req)
				if err != nil {
					return nil, fmt.Errorf("failed to get final response: %v", err)
				}
				defer finalResp.Body.Close()

				// Read final response body
				finalBytes, err := io.ReadAll(finalResp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read final response: %v", err)
				}

				if c.debug {
					fmt.Printf("Raw final response: %s\n", string(finalBytes))
				}

				// Check if the response is an error
				if finalResp.StatusCode != http.StatusOK {
					return nil, fmt.Errorf("request failed with status %d: %s", finalResp.StatusCode, string(finalBytes))
				}

				// Parse the response to get the audio URL
				var responseData struct {
					Audio struct {
						URL         string `json:"url"`
						ContentType string `json:"content_type"`
						FileName    string `json:"file_name"`
						FileSize    int    `json:"file_size"`
					} `json:"audio"`
					DurationMs int `json:"duration_ms"`
				}

				if err := json.Unmarshal(finalBytes, &responseData); err != nil {
					return nil, fmt.Errorf("failed to parse final response: %v", err)
				}

				if c.debug {
					fmt.Printf("Parsed Response: %+v\n", responseData)
				}

				// Check if we have a valid audio URL
				if responseData.Audio.URL == "" {
					return nil, fmt.Errorf("no audio URL in response")
				}

				// Create an AudioResponse to return
				audioResp := &AudioResponse{
					Audio: struct {
						URL         string `json:"url"`
						ContentType string `json:"content_type"`
						FileName    string `json:"file_name"`
						FileSize    int    `json:"file_size"`
					}{
						URL:         responseData.Audio.URL,
						ContentType: responseData.Audio.ContentType,
						FileName:    responseData.Audio.FileName,
						FileSize:    responseData.Audio.FileSize,
					},
					DurationMs: responseData.DurationMs,
				}

				return audioResp, nil
			case "FAILED":
				// Send failure message to user
				botValue := reflect.ValueOf(bot)
				sendPMMethod := botValue.MethodByName("SendPM")
				if sendPMMethod.IsValid() {
					sendPMMethod.Call([]reflect.Value{
						reflect.ValueOf(ctx),
						reflect.ValueOf(userNick),
						reflect.ValueOf("Your speech generation request failed."),
					})
				}
				return nil, fmt.Errorf("request failed")
			}
		}
	}
}

// makeRequest makes a request to the Fal.ai API
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	// Create request
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %v", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.apiKey)

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	return resp, nil
}

func (c *Client) GetModels(commandType string) (map[string]Model, error) {
	switch commandType {
	case "text2image":
		return Text2ImageModels, nil
	case "text2speech":
		return Text2SpeechModels, nil
	default:
		return nil, fmt.Errorf("unknown command type: %s", commandType)
	}
}

func (c *Client) GetCurrentModel(commandType string) (Model, error) {
	modelName, exists := GetDefaultModel(commandType)
	if !exists {
		return Model{}, fmt.Errorf("no default model found for command type: %s", commandType)
	}

	model, exists := GetModel(modelName, commandType)
	if !exists {
		return Model{}, fmt.Errorf("model not found: %s", modelName)
	}

	return model, nil
}

func (c *Client) SetCurrentModel(commandType, modelName string) error {
	if !SetDefaultModel(commandType, modelName) {
		return fmt.Errorf("failed to set default model %s for command type %s", modelName, commandType)
	}
	return nil
}
