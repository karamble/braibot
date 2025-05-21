package braibottypes

import (
	"context"
	"fmt"

	"github.com/companyzero/bisonrelay/zkidentity"
	kit "github.com/vctt94/bisonbotkit"
)

// BisonBotAdapter adapts *kit.Bot to the BotInterface
// This allows us to use *kit.Bot where BotInterface is required
// and provides the required SendPM, SendGC, and SendGCMessage methods
// with the correct signatures.
type BisonBotAdapter struct {
	bot *kit.Bot
}

func NewBisonBotAdapter(bot *kit.Bot) *BisonBotAdapter {
	return &BisonBotAdapter{bot: bot}
}

func (a *BisonBotAdapter) SendPM(ctx context.Context, uid zkidentity.ShortID, msg string) error {
	return a.bot.SendPM(ctx, uid.String(), msg)
}

func (a *BisonBotAdapter) SendGC(ctx context.Context, gc string, msg string) error {
	return a.bot.SendGC(ctx, gc, msg)
}

func (a *BisonBotAdapter) SendGCMessage(ctx context.Context, gc string, channel string, msg string) error {
	// *kit.Bot does not support channels, so we just send to the group
	return a.bot.SendGC(ctx, gc, msg)
}

// MessageSender provides a unified interface for sending messages in both PM and group chat contexts
type MessageSender struct {
	bot BotInterface
}

// NewMessageSender creates a new MessageSender instance
func NewMessageSender(bot BotInterface) *MessageSender {
	return &MessageSender{bot: bot}
}

// SendMessage sends a message to the user in the appropriate context
func (s *MessageSender) SendMessage(ctx context.Context, msgCtx MessageContext, message string) error {
	if msgCtx.IsPM {
		return s.bot.SendPM(ctx, msgCtx.Sender, message)
	}
	return s.bot.SendGC(ctx, msgCtx.GC, message)
}

// SendErrorMessage sends an error message to the user
func (s *MessageSender) SendErrorMessage(ctx context.Context, msgCtx MessageContext, err error) error {
	errorMsg := fmt.Sprintf("❌ Error: %v", err)
	return s.SendMessage(ctx, msgCtx, errorMsg)
}

// SendSuccessMessage sends a success message to the user
func (s *MessageSender) SendSuccessMessage(ctx context.Context, msgCtx MessageContext, message string) error {
	successMsg := fmt.Sprintf("✅ %s", message)
	return s.SendMessage(ctx, msgCtx, successMsg)
}
