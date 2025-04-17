// Copyright (c) 2025 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package faladapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// BotProgressCallback implements fal.ProgressCallback for sending updates to users via the bot.
type BotProgressCallback struct {
	bot      *kit.Bot
	userNick string

	// Throttling fields
	lastQueueUpdate    time.Time
	lastProgressUpdate time.Time
	lastLogMessage     time.Time
	lastSpecialMessage time.Time

	// Minimum interval between updates
	queueUpdateInterval    time.Duration
	progressUpdateInterval time.Duration
	logMessageInterval     time.Duration
	specialMessageInterval time.Duration

	// Latest messages
	latestQueueMessage    string
	latestProgressMessage string
	latestLogMessage      string

	// Track the last sent message to avoid duplicates
	lastSentMessage string
}

// NewBotProgressCallback creates a new BotProgressCallback with default throttling intervals.
func NewBotProgressCallback(bot *kit.Bot, userNick string) *BotProgressCallback {
	return &BotProgressCallback{
		bot:      bot,
		userNick: userNick,
		// Default intervals: 30 seconds for queue updates, 20 seconds for progress, 15 seconds for logs, 2 minutes for special messages
		queueUpdateInterval:    30 * time.Second,
		progressUpdateInterval: 20 * time.Second,
		logMessageInterval:     15 * time.Second,
		specialMessageInterval: 2 * time.Minute,
	}
}

// OnQueueUpdate sends queue position updates to the user with throttling.
func (c *BotProgressCallback) OnQueueUpdate(position int, eta time.Duration) {
	// Store the latest message
	c.latestQueueMessage = fmt.Sprintf("Queue position: %d, ETA: %v", position, eta)

	// Check if enough time has passed since the last update
	if time.Since(c.lastQueueUpdate) < c.queueUpdateInterval {
		return
	}

	c.bot.SendPM(context.Background(), c.userNick, c.latestQueueMessage)
	c.lastQueueUpdate = time.Now()
}

// OnProgress sends progress updates to the user with throttling.
func (c *BotProgressCallback) OnProgress(status string) {
	// Store the latest message
	c.latestProgressMessage = fmt.Sprintf("Status: %s", status)

	// Check if enough time has passed since the last update
	if time.Since(c.lastProgressUpdate) < c.progressUpdateInterval {
		return
	}

	// Send the progress message
	c.bot.SendPM(context.Background(), c.userNick, c.latestProgressMessage)

	// If status is IN_PROGRESS, send a special message about the expected processing time
	// but only once every 2 minutes at maximum
	if status == "IN_PROGRESS" && time.Since(c.lastSpecialMessage) >= c.specialMessageInterval {
		c.bot.SendPM(context.Background(), c.userNick, "The Video generation is in process\nVideo generation can take a long time, up to 20 minutes\nDuring the process the bot does not respond to any commands, please be patient")
		c.lastSpecialMessage = time.Now()
	}

	c.lastProgressUpdate = time.Now()
}

// OnError sends error messages to the user (no throttling for errors).
func (c *BotProgressCallback) OnError(err error) {
	c.bot.SendPM(context.Background(), c.userNick, fmt.Sprintf("Error: %v", err))
}

// OnLogMessage sends log messages to the user with throttling.
func (c *BotProgressCallback) OnLogMessage(message string) {
	var newMessage string

	// Check if the message is a JSON string containing logs
	if strings.Contains(message, `"logs":`) {
		// Parse the JSON response
		var response struct {
			Logs []struct {
				Message string `json:"message"`
				Labels  struct {
					LoggedAt string `json:"logged_at"`
				} `json:"labels"`
			} `json:"logs"`
		}

		if err := json.Unmarshal([]byte(message), &response); err != nil {
			// If JSON parsing fails, fall back to non-JSON handling
			lines := strings.Split(message, "\n")
			if len(lines) > 0 {
				lastLine := lines[len(lines)-1]
				newMessage = fmt.Sprintf("Log: %s", lastLine)
			} else {
				newMessage = fmt.Sprintf("Log: %s", message)
			}
		} else if len(response.Logs) > 0 {
			// Find the log entry with the latest logged_at timestamp
			var latestTime time.Time
			var latestMessage string

			for _, log := range response.Logs {
				t, err := time.Parse(time.RFC3339Nano, log.Labels.LoggedAt)
				if err != nil {
					continue
				}
				if t.After(latestTime) {
					latestTime = t
					latestMessage = log.Message
				}
			}

			if latestMessage != "" {
				newMessage = fmt.Sprintf("Progress: %s", latestMessage)
			}
		}
	} else {
		// For non-JSON messages, split into lines and take the last line
		lines := strings.Split(message, "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			newMessage = fmt.Sprintf("Log: %s", lastLine)
		} else {
			newMessage = fmt.Sprintf("Log: %s", message)
		}
	}

	// Always update the latest message
	c.latestLogMessage = newMessage

	// Check if enough time has passed since the last update
	if time.Since(c.lastLogMessage) < c.logMessageInterval {
		return
	}

	// Check if this is the same message we last sent
	if c.latestLogMessage == c.lastSentMessage {
		return
	}

	// Send the message
	c.bot.SendPM(context.Background(), c.userNick, c.latestLogMessage)
	c.lastLogMessage = time.Now()
	c.lastSentMessage = c.latestLogMessage
}

// GenerateImage generates an image using the fal package.
func GenerateImage(ctx context.Context, client *fal.Client, req fal.ImageRequest) (*fal.ImageResponse, error) {
	return client.GenerateImage(ctx, req)
}

// GenerateSpeech generates speech using the fal package.
func GenerateSpeech(ctx context.Context, client *fal.Client, text, voiceID string, bot *kit.Bot, userNick string) (*fal.AudioResponse, error) {
	// Create progress callback
	progress := NewBotProgressCallback(bot, userNick)

	// Create speech request
	req := fal.SpeechRequest{
		Text:     text,
		VoiceID:  voiceID,
		Options:  map[string]interface{}{},
		Progress: progress,
	}

	// Generate speech
	resp, err := client.GenerateSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate speech: %v", err)
	}

	return resp, nil
}

// GetModel returns a model by name and type.
func GetModel(name, modelType string) (fal.Model, bool) {
	return fal.GetModel(name, modelType)
}

// GetModels returns all available models for a command type.
func GetModels(commandType string) (map[string]fal.Model, bool) {
	return fal.GetModels(commandType)
}

// GetCurrentModel returns the current model for a command type.
func GetCurrentModel(commandType string) (fal.Model, bool) {
	return fal.GetCurrentModel(commandType)
}

// SetCurrentModel sets the current model for a command type.
func SetCurrentModel(commandType, modelName string) error {
	return fal.SetCurrentModel(commandType, modelName)
}

// GenerateImageFromImage generates an image from an existing image using the fal package.
func GenerateImageFromImage(ctx context.Context, client *fal.Client, prompt, imageURL, modelName string, progress fal.ProgressCallback) (*fal.ImageResponse, error) {
	req := fal.ImageRequest{
		Prompt: prompt,
		Model:  modelName,
		Options: map[string]interface{}{
			"image_url": imageURL,
		},
		Progress: progress,
	}

	return client.GenerateImage(ctx, req)
}
