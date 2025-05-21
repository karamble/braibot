package commands

import (
	"context"

	braibottypes "github.com/karamble/braibot/internal/types"
)

// Command is an alias for types.Command
type Command = braibottypes.Command

// MessageContext is an alias for types.MessageContext
type MessageContext = braibottypes.MessageContext

// ReplyFunc is a function type that handles sending replies in both PM and group chat contexts
type ReplyFunc func(ctx context.Context, message string) error
