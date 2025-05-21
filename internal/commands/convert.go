package commands

import (
	"context"

	"github.com/companyzero/bisonrelay/zkidentity"
	braibottypes "github.com/karamble/braibot/internal/types"
	"github.com/vctt94/bisonbotkit"
)

// NewReplyFunc creates a new ReplyFunc based on the message context and bot instance
func NewReplyFunc(bot *bisonbotkit.Bot, msgCtx braibottypes.MessageContext) ReplyFunc {
	return func(ctx context.Context, message string) error {
		if !msgCtx.IsPM {
			return bot.SendGC(ctx, msgCtx.GC, message)
		}
		return bot.SendPM(ctx, msgCtx.Sender.String(), message)
	}
}

// ConvertPMToMessageContext converts a ReceivedPM to a MessageContext
func ConvertPMToMessageContext(pm braibottypes.ReceivedPM) braibottypes.MessageContext {
	var sender zkidentity.ShortID
	copy(sender[:], pm.Uid)
	return braibottypes.MessageContext{
		Nick:    pm.Nick,
		Uid:     pm.Uid,
		Message: pm.Msg.Message,
		IsPM:    true,
		Sender:  sender,
	}
}

// ConvertGCToMessageContext converts a GCReceivedMsg to a MessageContext
func ConvertGCToMessageContext(gc braibottypes.GCReceivedMsg) braibottypes.MessageContext {
	var sender zkidentity.ShortID
	copy(sender[:], gc.Uid)
	return braibottypes.MessageContext{
		Nick:    gc.Nick,
		Uid:     gc.Uid,
		Message: gc.Msg.Message,
		IsPM:    false,
		GC:      gc.GcAlias,
		Sender:  sender,
	}
}
