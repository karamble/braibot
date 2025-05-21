package braibottypes

import (
	"context"

	"github.com/companyzero/bisonrelay/zkidentity"
)

// BotInterface defines the interface for bot operations
type BotInterface interface {
	SendPM(ctx context.Context, uid zkidentity.ShortID, msg string) error
	SendGC(ctx context.Context, gc string, msg string) error
	SendGCMessage(ctx context.Context, gc string, channel string, msg string) error
}

// DBManagerInterface defines the interface for database operations
type DBManagerInterface interface {
	GetBalance(userID string) (int64, error)
	UpdateBalance(userID string, amount int64) error
	Close() error
}
