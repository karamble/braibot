package utils

import (
	"context"

	kit "github.com/vctt94/bisonbotkit"
)

// SendToUser sends a message to either a PM or a group chat based on isPM.
func SendToUser(ctx context.Context, bot *kit.Bot, isPM bool, nick, gc, msg string) error {
	if isPM {
		return bot.SendPM(ctx, nick, msg)
	}
	return bot.SendGC(ctx, gc, msg)
}
