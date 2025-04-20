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
	Category    string
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
	categories := make(map[string][]Command)
	categoryOrder := []string{"🎯 Basic", "🔧 Model Configuration", "🎨 AI Generation"}

	// Group commands by category
	for _, cmd := range r.commands {
		cat := cmd.Category
		if cat == "" {
			cat = "Other" // Default category if none provided
		}
		categories[cat] = append(categories[cat], cmd)
	}

	var helpMsg strings.Builder

	// Function to append category section to help message
	appendCategory := func(categoryName string) {
		if cmds, ok := categories[categoryName]; ok {
			helpMsg.WriteString(fmt.Sprintf("\n## %s\n", categoryName))
			helpMsg.WriteString("| Command | Description |\n| -------- | ----------- |\n")
			for _, cmd := range cmds {
				helpMsg.WriteString(fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description))
			}
			delete(categories, categoryName) // Remove processed category
		}
	}

	// Append categories in the specified order
	for _, catName := range categoryOrder {
		appendCategory(catName)
	}

	// Append any remaining categories (e.g., "Other")
	for catName := range categories {
		appendCategory(catName)
	}

	return helpMsg.String()
}
