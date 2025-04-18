package commands

import (
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/image"
	"github.com/karamble/braibot/internal/speech"
	"github.com/karamble/braibot/internal/video"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// InitializeCommands creates and registers all available commands
func InitializeCommands(dbManager *database.DBManager, cfg *config.BotConfig, bot *kit.Bot, debug bool) *Registry {
	registry := NewRegistry()

	// Create Fal client (assuming API key is in extra config)
	falClient := fal.NewClient(cfg.ExtraConfig["falapikey"], fal.WithDebug(debug))

	// Create Services
	imageService := image.NewImageService(falClient, dbManager, bot, debug)
	videoService := video.NewVideoService(falClient, dbManager, bot, debug)
	speechService := speech.NewSpeechService(falClient, dbManager, bot, debug)

	// Register help command
	registry.Register(HelpCommand(registry, dbManager))

	// Register model-related commands
	registry.Register(ListModelsCommand())
	registry.Register(SetModelCommand(registry))

	// Register AI commands (using services)
	registry.Register(Text2ImageCommand(dbManager, imageService, debug))
	registry.Register(Text2SpeechCommand(dbManager, speechService, debug))
	registry.Register(Image2ImageCommand(dbManager, imageService, debug))
	registry.Register(Image2VideoCommand(dbManager, videoService, debug))
	registry.Register(Text2VideoCommand(dbManager, videoService, debug))

	// Register balance and rate commands
	registry.Register(BalanceCommand(dbManager, debug))
	registry.Register(RateCommand())

	return registry
}
