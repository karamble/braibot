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

	// Get billing enabled flag from config (defaulting to true)
	billingEnabledStr := cfg.ExtraConfig["billingenabled"] // Already validated in config check
	billingEnabled := (billingEnabledStr == "true")

	// Create Services, passing the billing flag
	imageService := image.NewImageService(falClient, dbManager, bot, debug, billingEnabled)
	videoService := video.NewVideoService(falClient, dbManager, bot, debug, billingEnabled)    // Assuming NewVideoService signature is updated
	speechService := speech.NewSpeechService(falClient, dbManager, bot, debug, billingEnabled) // Assuming NewSpeechService signature is updated

	// Register help command
	registry.Register(HelpCommand(registry, dbManager))

	// Register model-related commands
	registry.Register(ListModelsCommand())
	registry.Register(SetModelCommand(registry))

	// Register AI commands (using services)
	// Pass the billingEnabled flag to commands that might need it directly (like balance)
	registry.Register(Text2ImageCommand(dbManager, imageService, debug))
	registry.Register(Text2SpeechCommand(dbManager, speechService, debug)) // Pass service instance
	registry.Register(Image2ImageCommand(dbManager, imageService, debug))
	registry.Register(Image2VideoCommand(dbManager, videoService, debug))
	registry.Register(Text2VideoCommand(dbManager, videoService, debug))
	registry.Register(AICommand(debug))

	// Register balance and rate commands
	registry.Register(BalanceCommand(dbManager, debug, billingEnabled)) // Pass billing flag
	registry.Register(RateCommand())

	return registry
}
