# Developer Guide

Welcome to the BraiBot Developer Guide! This section is for contributors and advanced users who want to understand, extend, or contribute to the project.

## Table of Contents

1. [Project Structure](#project-structure)
2. [Setup](#setup)
3. [Bison Relay Integration](#bison-relay-integration)
4. [fal.ai Integration](#falai-integration)
5. [n8n Integration](#n8n-integration)
6. [Adding Commands](#adding-commands)
7. [Adding Models](#adding-models)
8. [Contributing](#contributing)

## Project Structure

Overview of the codebase organization and key components.

### Database Implementation

The bot uses SQLite for local balance management:

#### Database Schema
```sql
CREATE TABLE IF NOT EXISTS user_balances (
    uid TEXT PRIMARY KEY,
    balance INTEGER NOT NULL DEFAULT 0
)
```

#### Key Components

1. **DBManager** (`internal/database/db.go`)
   - Handles database operations
   - Thread-safe with mutex locks
   - Manages SQLite connection
   - Provides balance operations

2. **Balance Management** (`internal/database/balance.go`)
   - `CheckAndDeductBalance`: Verifies and deducts costs
   - `GetUserBalance`: Retrieves user balance
   - Handles DCR atom conversions

3. **Currency Conversion** (`internal/database/currency.go`)
   - `USDToDCR`: Converts USD to DCR
   - Fetches exchange rates from CoinGecko
   - Caches rates for 5 minutes

#### Usage Example
```go
// Initialize database
dbManager, err := database.NewDBManager(appRoot)
if err != nil {
    log.Fatal(err)
}
defer dbManager.Close()

// Check and deduct balance
success, err := dbManager.CheckAndDeductBalance(userID, costUSD, debug)
if err != nil {
    // Handle error
}

// Get user balance
balance, err := dbManager.GetUserBalance(userID)
if err != nil {
    // Handle error
}
```

## Setup

How to set up a development environment for BraiBot.

## Bison Relay Integration

BraiBot is built on top of [bisonbotkit](https://github.com/vctt94/bisonbotkit), a Go framework for developing Bison Relay bots. This framework provides the core functionality for:
- RPC client management
- Message handling
- User authentication
- Payment processing
- Configuration management

The bisonbotkit framework handles the low-level Bison Relay protocol interactions, allowing BraiBot to focus on implementing its AI features and business logic.

Details on how BraiBot interacts with the Bison Relay platform:

## fal.ai Integration

BraiBot uses [fal.ai](https://fal.ai) for its generative AI features. To enable these features, you need:

1. A registered account on fal.ai
2. Funded credits in your fal.ai account
3. A valid API key

The following features require fal.ai integration:
- Text to Image generation
- Text to Video generation
- Text to Speech conversion
- Image to Image transformation
- Image to Video conversion

The API key is configured during the first launch of BraiBot through the setup wizard. You can also update it later in the configuration file.

## n8n Integration

The `!ai` command in BraiBot uses n8n webhooks for advanced AI workflows. You have two options to set up n8n:

1. **n8n.io Cloud Service**
   * Register at [n8n.io](https://n8n.io)
   * Create a workflow with a webhook trigger
   * Configure the webhook URL in BraiBot's config

2. **Self-hosted n8n** (Recommended for privacy and control)
   * Use [local-ai-packaged](https://github.com/coleam00/local-ai-packaged) - a comprehensive Docker setup that includes:
     - n8n for workflow automation
     - Ollama for local LLM support
     - PostgreSQL and Supabase for data storage
     - Additional AI tools and utilities
   * The package provides a complete local AI stack that can be run in a single Docker container
   * Perfect for privacy-focused deployments and custom AI workflows
   * Read more about [n8n-Integration](n8n-Integration)

To enable the `!ai` command:
1. Set up n8n (either cloud or self-hosted)
2. Create your AI workflow
3. Configure the webhook URL in BraiBot's config
4. Enable webhook functionality during first launch or in the config file

## Adding Commands

BraiBot uses a command registry system to manage bot commands. Here's how to add new commands:

### Command Structure

Each command is defined by a `Command` struct with the following fields:
```go
type Command struct {
    Name        string          // Command name (without !)
    Description string          // Help text description
    Category    string          // Category for help menu
    Handler     CommandFunc     // Command implementation
}
```

### Creating a New Command

1. Create a new file in `internal/commands/` (e.g., `mycommand.go`)
2. Implement your command handler:
```go
package commands

import (
    "context"
    "github.com/karamble/braibot/internal/types"
)

func MyCommand() types.Command {
    return types.Command{
        Name:        "mycommand",
        Description: "Description of what my command does",
        Category:    "🎯 Basic",  // Choose from: "🎯 Basic", "🔧 Model Configuration", "🎨 AI Generation"
        Handler: types.CommandFunc(func(ctx context.Context, msgCtx types.MessageContext, args []string, sender *types.MessageSender, db types.DBManagerInterface) error {
            // Command implementation
            return sender.SendMessage(ctx, msgCtx, "Command response")
        }),
    }
}
```

### Command Features

- **Message Context**: Access to message details (sender, PM/GC, etc.)
- **Arguments**: Command arguments are passed as a string slice
- **Message Sender**: Helper for sending responses
- **Database Access**: Interface for balance and other DB operations

### Best Practices

1. **Error Handling**
   - Use `sender.SendErrorMessage()` for error responses
   - Return errors for logging and debugging

2. **PM vs GC**
   - Check `msgCtx.IsPM` for private messages
   - Use appropriate response methods

3. **Balance Management**
   - Use `db.CheckAndDeductBalance()` for paid commands
   - Handle insufficient balance errors

4. **Testing**
   - Add tests in `command_test.go`
   - Test both success and error cases

### Example Command

Here's a complete example of a balance command:
```go
func BalanceCommand() types.Command {
    return types.Command{
        Name:        "balance",
        Description: "💰 Show your current balance",
        Category:    "🎯 Basic",
        Handler: types.CommandFunc(func(ctx context.Context, msgCtx types.MessageContext, args []string, sender *types.MessageSender, db types.DBManagerInterface) error {
            // Only respond in private messages
            if !msgCtx.IsPM {
                return nil
            }

            userID := string(msgCtx.Uid)
            balance, err := db.GetBalance(userID)
            if err != nil {
                return sender.SendErrorMessage(ctx, msgCtx, fmt.Errorf("failed to get balance: %v", err))
            }
            balanceMsg := fmt.Sprintf("💰 Your Balance: %d atoms", balance)
            return sender.SendMessage(ctx, msgCtx, balanceMsg)
        }),
    }
}
```

### Registering Commands

1. Add your command to `internal/commands/init.go`:
```go
func InitializeCommands(dbManager types.DBManagerInterface, cfg *config.BotConfig, bot *kit.Bot, debug bool) *Registry {
    registry := NewRegistry()
    // ... other commands ...
    registry.Register(MyCommand())
    return registry
}
```

2. Update the help command if needed (in `help.go`)

## Adding Models and fal.ai endpoints

This guide explains how to integrate new fal.ai endpoints into BraiBot. The integration process involves several steps to ensure proper type safety, validation, and error handling.

### Model Types

1. **Base Request Types**
   - `BaseImageRequest`: For image generation/transformation
   - `BaseVideoRequest`: For video generation
   - `BaseSpeechRequest`: For text-to-speech

2. **Model Options Methods**
   ```go
   // Your options struct should implement these methods:
   func (o *MyModelOptions) GetDefaultValues() map[string]interface{} { /* ... */ }
   func (o *MyModelOptions) Validate() error { /* ... */ }
   ```

### Integration Steps

1. **Define Model Options**
   ```go
   type MyModelOptions struct {
       // Add model-specific options
       Option1 string  `json:"option1,omitempty"`
       Option2 float64 `json:"option2,omitempty"`
   }

   func (o *MyModelOptions) GetDefaultValues() map[string]interface{} {
       return map[string]interface{}{
           "option1": "default_value",
           "option2": 1.0,
       }
   }

   func (o *MyModelOptions) Validate() error {
       // Add validation logic
       return nil
   }
   ```

2. **Create Request Type**
   ```go
   type MyModelRequest struct {
       BaseImageRequest // or BaseVideoRequest or BaseSpeechRequest
       // Add model-specific fields
       Option1 string  `json:"option1,omitempty"`
       Option2 float64 `json:"option2,omitempty"`
   }
   ```

3. **Add Model Definition**
   ```go
   func MyModel() Model {
       return Model{
           Name:        "my-model",
           Description: "Description of the model",
           PriceUSD:    0.01,
           Type:        "image", // or "video" or "speech"
           HelpDoc:     "Help documentation for the model",
           Options:     &MyModelOptions{},
       }
   }
   ```

4. **Register Model**
   Add your model to the appropriate model registry in:
   - `text_image_models.go` for text-to-image
   - `text_video_models.go` for text-to-video
   - `text_speech_models.go` for text-to-speech
   - `image_image_models.go` for image-to-image
   - `image_video_models.go` for image-to-video

### Example: Adding a New Image Model

```go
// 1. Define options
type MyImageOptions struct {
    ImageSize string  `json:"image_size,omitempty"`
    Quality   float64 `json:"quality,omitempty"`
}

func (o *MyImageOptions) GetDefaultValues() map[string]interface{} {
    return map[string]interface{}{
        "image_size": "1024x1024",
        "quality":    0.8,
    }
}

func (o *MyImageOptions) Validate() error {
    validSizes := map[string]bool{
        "1024x1024": true,
        "512x512":   true,
    }
    if !validSizes[o.ImageSize] {
        return fmt.Errorf("invalid image size: %s", o.ImageSize)
    }
    if o.Quality < 0 || o.Quality > 1 {
        return fmt.Errorf("quality must be between 0 and 1")
    }
    return nil
}

// 2. Create request type
type MyImageRequest struct {
    BaseImageRequest
    ImageSize string  `json:"image_size,omitempty"`
    Quality   float64 `json:"quality,omitempty"`
}

// 3. Add model definition
func MyImageModel() Model {
    return Model{
        Name:        "my-image-model",
        Description: "A custom image generation model",
        PriceUSD:    0.01,
        Type:        "image",
        HelpDoc:     "Generate images with custom size and quality settings",
        Options:     &MyImageOptions{},
    }
}

// 4. Register model
func init() {
    RegisterTextImageModel(MyImageModel())
}
```

### Progress Tracking

The fal.ai client supports progress tracking through the `ProgressCallback` interface:

```go
type ProgressCallback interface {
    OnQueueUpdate(position int, eta time.Duration)
    OnLogMessage(message string)
    OnProgress(status string)
    OnError(err error)
}
```

Implement this interface to track:
- Queue position and ETA
- Processing status
- Error messages
- Log messages

### Error Handling

The client provides structured error handling:

```go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

Handle errors appropriately in your implementation:
1. Check for API errors
2. Validate responses
3. Provide user-friendly error messages

## Contributing

How to contribute to the project, coding standards, and submitting pull requests. 