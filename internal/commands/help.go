package commands

import (
	"context"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// HelpCommand returns the help command
func HelpCommand(registry *Registry) Command {
	return Command{
		Name:        "help",
		Description: "Shows this help message",
		Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
			return bot.SendPM(ctx, pm.Nick, registry.FormatHelpMessage())
		},
	}
}
