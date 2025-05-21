package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// MessageSender handles sending messages in both PM and GC contexts
type MessageSender struct {
	bot braibottypes.BotInterface
}

// NewMessageSender creates a new message sender
func NewMessageSender(bot braibottypes.BotInterface) *MessageSender {
	return &MessageSender{
		bot: bot,
	}
}

// SendMessage sends a message in the appropriate context
func (s *MessageSender) SendMessage(ctx context.Context, msgCtx braibottypes.MessageContext, msg string) error {
	if msgCtx.IsPM {
		return s.bot.SendPM(ctx, msgCtx.Sender, msg)
	}
	return s.bot.SendGC(ctx, msgCtx.GC, msg)
}

// SendErrorMessage sends an error message in the appropriate context
func (s *MessageSender) SendErrorMessage(ctx context.Context, msgCtx braibottypes.MessageContext, err error) error {
	msg := fmt.Sprintf("❌ Error: %v", err)
	return s.SendMessage(ctx, msgCtx, msg)
}

// SimpleProgressCallback implements fal.ProgressCallback for simple progress updates
type SimpleProgressCallback struct {
	bot         *braibottypes.BisonBotAdapter
	userNick    string
	commandName string
}

// NewSimpleProgressCallback creates a new simple progress callback function
func NewSimpleProgressCallback(bot *kit.Bot, userNick string, commandName string) fal.ProgressCallback {
	return &SimpleProgressCallback{
		bot:         braibottypes.NewBisonBotAdapter(bot),
		userNick:    userNick,
		commandName: commandName,
	}
}

// OnQueueUpdate implements fal.ProgressCallback
func (c *SimpleProgressCallback) OnQueueUpdate(position int, eta time.Duration) {
	msg := fmt.Sprintf("Queue position: %d, ETA: %v", position, eta)
	var userID zkidentity.ShortID
	c.bot.SendPM(context.Background(), userID, msg)
}

// OnLogMessage implements fal.ProgressCallback
func (c *SimpleProgressCallback) OnLogMessage(message string) {
	msg := fmt.Sprintf("Log: %s", message)
	var userID zkidentity.ShortID
	c.bot.SendPM(context.Background(), userID, msg)
}

// OnProgress implements fal.ProgressCallback
func (c *SimpleProgressCallback) OnProgress(status string) {
	msg := fmt.Sprintf("Progress for %s: %s", c.commandName, status)
	var userID zkidentity.ShortID
	c.bot.SendPM(context.Background(), userID, msg)
}

// OnError implements fal.ProgressCallback
func (c *SimpleProgressCallback) OnError(err error) {
	msg := fmt.Sprintf("Error in %s: %v", c.commandName, err)
	var userID zkidentity.ShortID
	c.bot.SendPM(context.Background(), userID, msg)
}

// CommandErrorCallback is a function type that handles error updates
type CommandErrorCallback func(err error) error

// NewCommandErrorCallback creates a new error callback function
func NewCommandErrorCallback(bot *braibottypes.BisonBotAdapter, userNick string, commandName string) CommandErrorCallback {
	return func(err error) error {
		// Format the error message
		errorMsg := fmt.Sprintf("Error in %s: %v", commandName, err)
		// Send the error message to the user
		var userID zkidentity.ShortID
		return bot.SendPM(context.Background(), userID, errorMsg)
	}
}
