package braibottypes

import (
	"context"
)

// Command represents a bot command
type Command struct {
	Name        string
	Description string
	Category    string
	Handler     CommandHandler
}

// CommandHandler defines the interface for command handlers
type CommandHandler interface {
	Handle(ctx context.Context, msgCtx MessageContext, args []string, sender *MessageSender, db DBManagerInterface) error
}

// CommandFunc is a function type that implements CommandHandler
type CommandFunc func(ctx context.Context, msgCtx MessageContext, args []string, sender *MessageSender, db DBManagerInterface) error

// Handle implements the CommandHandler interface for CommandFunc
func (f CommandFunc) Handle(ctx context.Context, msgCtx MessageContext, args []string, sender *MessageSender, db DBManagerInterface) error {
	return f(ctx, msgCtx, args, sender, db)
}
