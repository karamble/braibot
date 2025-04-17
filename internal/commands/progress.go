package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	kit "github.com/vctt94/bisonbotkit"
)

// CommandProgressCallback implements fal.ProgressCallback for sending updates to users via the bot.
type CommandProgressCallback struct {
	bot      *kit.Bot
	userNick string
	cmdType  string

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

// NewCommandProgressCallback creates a new CommandProgressCallback with default throttling intervals.
func NewCommandProgressCallback(bot *kit.Bot, userNick, cmdType string) *CommandProgressCallback {
	return &CommandProgressCallback{
		bot:      bot,
		userNick: userNick,
		cmdType:  cmdType,
		// Default intervals: 30 seconds for queue updates, 20 seconds for progress, 15 seconds for logs, 2 minutes for special messages
		queueUpdateInterval:    30 * time.Second,
		progressUpdateInterval: 20 * time.Second,
		logMessageInterval:     15 * time.Second,
		specialMessageInterval: 2 * time.Minute,
	}
}

// OnQueueUpdate sends queue position updates to the user with throttling.
func (c *CommandProgressCallback) OnQueueUpdate(position int, eta time.Duration) {
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
func (c *CommandProgressCallback) OnProgress(status string) {
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
		var message string
		switch c.cmdType {
		case "image2video":
			message = "The video generation is in process\nVideo generation can take a long time, up to 20 minutes\nDuring the process the bot does not respond to any commands, please be patient"
		case "image2image":
			message = "The image generation is in process\nImage generation can take a few minutes\nDuring the process the bot does not respond to any commands, please be patient"
		case "text2image":
			message = "The image generation is in process\nImage generation can take a few minutes\nDuring the process the bot does not respond to any commands, please be patient"
		case "text2speech":
			message = "The speech generation is in process\nSpeech generation can take a few minutes\nDuring the process the bot does not respond to any commands, please be patient"
		default:
			message = "The generation is in process\nThis may take a few minutes\nDuring the process the bot does not respond to any commands, please be patient"
		}
		c.bot.SendPM(context.Background(), c.userNick, message)
		c.lastSpecialMessage = time.Now()
	}

	c.lastProgressUpdate = time.Now()
}

// OnError sends error messages to the user (no throttling for errors).
func (c *CommandProgressCallback) OnError(err error) {
	c.bot.SendPM(context.Background(), c.userNick, fmt.Sprintf("Error: %v", err))
}

// OnLogMessage sends log messages to the user with throttling.
func (c *CommandProgressCallback) OnLogMessage(message string) {
	// For non-JSON messages, split into lines and take the last line
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		lastLine := lines[len(lines)-1]
		c.latestLogMessage = fmt.Sprintf("Log: %s", lastLine)
	} else {
		c.latestLogMessage = fmt.Sprintf("Log: %s", message)
	}

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
