package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
)

// Command represents a bot command
type Command struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error
}

// Registry holds all available commands
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name] = cmd
}

// Get returns a command by name
func (r *Registry) Get(name string) (Command, bool) {
	cmd, exists := r.commands[name]
	return cmd, exists
}

// GetAll returns all registered commands
func (r *Registry) GetAll() map[string]Command {
	return r.commands
}

// IsCommand checks if a message is a command (starts with !)
func IsCommand(msg string) (string, []string, bool) {
	if !strings.HasPrefix(msg, "!") {
		return "", nil, false
	}

	parts := strings.Fields(msg[1:]) // Remove ! and split
	if len(parts) == 0 {
		return "", nil, false
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	return cmd, args, true
}

// FormatHelpMessage formats a help message for all registered commands
func (r *Registry) FormatHelpMessage() string {
	// Define command categories
	universalCommands := []string{"help", "balance", "rate"}
	modelCommands := []string{"listmodels", "setmodel"}
	generationCommands := []string{"text2image", "image2image", "image2video", "text2speech"}

	// Create help message with sections
	helpMsg := "## 🎯 Basic Commands\n"
	helpMsg += "| Command | Description |\n| -------- | ----------- |\n"
	for _, cmdName := range universalCommands {
		if cmd, exists := r.commands[cmdName]; exists {
			helpMsg += fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description)
		}
	}

	helpMsg += "\n## 🔧 Model Configuration\n"
	helpMsg += "| Command | Description |\n| -------- | ----------- |\n"
	for _, cmdName := range modelCommands {
		if cmd, exists := r.commands[cmdName]; exists {
			helpMsg += fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description)
		}
	}

	helpMsg += "\n## 🎨 AI Generation\n"
	helpMsg += "| Command | Description |\n| -------- | ----------- |\n"
	for _, cmdName := range generationCommands {
		if cmd, exists := r.commands[cmdName]; exists {
			helpMsg += fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description)
		}
	}

	return helpMsg
}
