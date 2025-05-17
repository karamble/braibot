package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	braibottypes "github.com/karamble/braibot/internal/types"
	kit "github.com/vctt94/bisonbotkit"
)

// CommandProgressCallback implements fal.ProgressCallback for sending updates to users via the bot.
type CommandProgressCallback struct {
	bot      *braibottypes.BisonBotAdapter
	userNick string
	userID   zkidentity.ShortID
	cmdType  string
	isPM     bool
	gc       string

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

	// Track the last sent message to avoid duplicates for each type
	lastSentMessage         string // Used by OnLogMessage
	lastSentQueueMessage    string // Added for OnQueueUpdate
	lastSentProgressMessage string // Added for OnProgress
}

// NewCommandProgressCallback creates a new CommandProgressCallback with default throttling intervals.
func NewCommandProgressCallback(bot *kit.Bot, userNick string, userID zkidentity.ShortID, cmdType string, isPM bool, gc string) *CommandProgressCallback {
	return &CommandProgressCallback{
		bot:      braibottypes.NewBisonBotAdapter(bot),
		userNick: userNick,
		userID:   userID,
		cmdType:  cmdType,
		isPM:     isPM,
		gc:       gc,
		// Default intervals: 30 seconds for queue updates, 20 seconds for progress, 15 seconds for logs, 2 minutes for special messages
		queueUpdateInterval:    30 * time.Second,
		progressUpdateInterval: 20 * time.Second,
		logMessageInterval:     15 * time.Second,
		specialMessageInterval: 2 * time.Minute,
	}
}

// sendMessage sends a message to the appropriate channel based on the message context
func (c *CommandProgressCallback) sendMessage(msg string) {
	if c.isPM {
		c.bot.SendPM(context.Background(), c.userID, msg)
	} else {
		c.bot.SendGC(context.Background(), c.gc, msg)
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

	// Check if this is the same message we last sent for queue updates
	if c.latestQueueMessage == c.lastSentQueueMessage {
		return
	}

	c.sendMessage(c.latestQueueMessage)
	c.lastQueueUpdate = time.Now()
	c.lastSentQueueMessage = c.latestQueueMessage // Update last sent queue message
}

// OnProgress sends progress updates to the user with throttling.
func (c *CommandProgressCallback) OnProgress(status string) {
	// Store the latest message
	c.latestProgressMessage = fmt.Sprintf("Status: %s", status)

	// Check if enough time has passed since the last update
	if time.Since(c.lastProgressUpdate) < c.progressUpdateInterval {
		return
	}

	// Check if this is the same message we last sent for progress updates
	if c.latestProgressMessage == c.lastSentProgressMessage {
		return
	}

	// Send the progress message
	c.sendMessage(c.latestProgressMessage)
	c.lastProgressUpdate = time.Now()
	c.lastSentProgressMessage = c.latestProgressMessage // Update last sent progress message

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
		c.sendMessage(message)
		c.lastSpecialMessage = time.Now()
	}
}

// OnError sends error messages to the user (no throttling for errors).
func (c *CommandProgressCallback) OnError(err error) {
	c.sendMessage(fmt.Sprintf("Error: %v", err))
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
	c.sendMessage(c.latestLogMessage)
	c.lastLogMessage = time.Now()
	c.lastSentMessage = c.latestLogMessage
}
