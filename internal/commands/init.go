package commands

import (
	"github.com/karamble/braibot/internal/database"
)

// InitializeCommands creates and registers all available commands
func InitializeCommands(dbManager *database.DBManager, debug bool) *Registry {
	registry := NewRegistry()

	// Register help command
	registry.Register(HelpCommand(registry))

	// Register model-related commands
	registry.Register(ListModelsCommand())
	registry.Register(SetModelCommand(registry))

	// Register AI commands
	registry.Register(Text2ImageCommand(dbManager, debug))
	registry.Register(Text2SpeechCommand(dbManager, debug))
	registry.Register(Image2ImageCommand(dbManager, debug))

	// Register balance and rate commands
	registry.Register(BalanceCommand(dbManager, debug))
	registry.Register(RateCommand())

	return registry
}
