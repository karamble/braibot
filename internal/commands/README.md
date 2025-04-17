# Commands Package

This package provides command handlers for the Braibot, a chatbot for the Bison Relay platform. The commands package implements various AI-powered features using the Fal.ai API.

## Overview

The commands package contains handlers for different bot commands that users can invoke through private messages. Each command is implemented as a function that returns a `Command` struct with a name, description, and handler function.

## Integration Guide

### Package Structure

The commands package is part of a larger system with the following structure:

```
braibot/
├── pkg/
│   └── fal/           # Core Fal.ai API client
├── internal/
│   ├── commands/      # Command handlers
│   ├── faladapter/    # Adapter for Fal.ai integration
│   └── database/      # User balance management
```

### Prerequisites

1. **BisonBotKit Setup**
   - A working BisonBotKit project
   - Bison Relay client running with RPC enabled
   - Proper configuration for BisonBotKit

2. **Required Dependencies**
   - Fal.ai API key
   - A Bison Relay Client running with json rpc interface enabled


### Quick Start

1. **Add Dependencies**
   ```go
   import (
       "github.com/karamble/braibot/internal/commands"    // Command handlers
       "github.com/karamble/braibot/internal/faladapter"  // Fal.ai adapter
       "github.com/karamble/braibot/internal/database"    // Database management
       "github.com/karamble/braibot/pkg/fal"             // Fal.ai client
       "github.com/vctt94/bisonbotkit"                   // BisonBotKit
       "github.com/vctt94/bisonbotkit/config"            // BisonBotKit config
       "github.com/vctt94/bisonbotkit/logging"           // BisonBotKit logging
       "github.com/vctt94/bisonbotkit/utils"             // BisonBotKit utilities
   )
   ```

2. **Initialize Database and Bot**
   ```go
   // Get the BisonBotKit app root path
   appRoot := utils.CleanAndExpandPath(*flagAppRoot)
   
   // Initialize database manager
   dbManager, err := database.NewDBManager(appRoot)
   if err != nil {
       log.Fatal(err)
   }
   defer dbManager.Close()

   // Initialize logging
   logBackend, err := logging.NewLogBackend(logging.LogConfig{
       LogFile:        filepath.Join(appRoot, "logs", "bot.log"),
       DebugLevel:     "info",
       MaxLogFiles:    5,
       MaxBufferLines: 1000,
   })
   if err != nil {
       log.Fatal(err)
   }
   defer logBackend.Close()

   // Load bot configuration
   cfg, err := botkitconfig.LoadBotConfig(appRoot, "bot.conf")
   if err != nil {
       log.Fatal(err)
   }

   // Create new bot instance
   bot, err := kit.NewBot(cfg, logBackend)
   if err != nil {
       log.Fatal(err)
   }
   ```

3. **Initialize Commands**
   ```go
   // Create command registry with debug mode (optional)
   // This automatically registers all available commands:
   // - Basic commands: help, balance, rate
   // - Model configuration: listmodels, setmodel
   // - AI commands: text2image, text2speech, image2image, image2video
   commandRegistry := commands.InitializeCommands(dbManager, debug)
   ```

4. **Set Up Message Handler**
   ```go
   // Create a channel for receiving messages
   pmChan := make(chan types.ReceivedPM)
   cfg.PMChan = pmChan

   // Add a goroutine to handle messages
   go func() {
       for pm := range pmChan {
           // Check if the message is a command
           if cmd, args, isCmd := commands.IsCommand(pm.Msg.Message); isCmd {
               if command, exists := commandRegistry.Get(cmd); exists {
                   if err := command.Handler(context.Background(), bot, cfg, pm, args); err != nil {
                       log.Warnf("Error executing command %s: %v", cmd, err)
                   }
               }
           }
       }
   }()
   ```

### Configuration

Add the following to your bot's configuration:

```go
cfg := &config.BotConfig{
    ExtraConfig: map[string]string{
        "falapikey": "your-fal-ai-api-key",
    },
}
```

### Database Setup

The commands package includes an integrated SQLite3 database for managing user balances. The database is automatically created and initialized when you create a new DBManager:

```go
dbManager, err := database.NewDBManager("path/to/app/root")
```

The database will be created at `path/to/app/root/data/balances.db` with the following features:
- Automatic table creation and schema management
- Thread-safe operations with mutex locking
- Automatic user balance initialization
- Built-in transaction support

### Lightning Network Integration

The commands package integrates with Decred's Lightning Network for processing payments. Users can send tips through Bison Relay to add funds to their balance that is stored in the database for each user.

1. **Balance Management**
   - Users can check their balance with `!balance`
   - Balances are stored in DCR (Decred)
   - Automatic conversion between USD and DCR for AI operations

2. **Payment Processing**
   - Automatic deduction of balance for AI operations
   - Real-time balance updates
   - Transaction history tracking

### Error Handling

The commands package includes comprehensive error handling:

```go
if err != nil {
    // Common error types:
    switch {
    case errors.Is(err, database.ErrInsufficientBalance):
        // Handle insufficient balance
    case errors.Is(err, faladapter.ErrInvalidModel):
        // Handle invalid model selection
    case errors.Is(err, faladapter.ErrAPIError):
        // Handle Fal.ai API errors
    default:
        // Handle other errors
    }
}
```

### Debug Mode

Enable debug mode for detailed logging:

```go
// Initialize commands with debug mode
registry := commands.InitializeCommands(dbManager, true)
```

Debug mode provides:
- Detailed error messages
- API request/response logging
- Balance operation logging
- Model selection tracking

### Customizing and Extending

#### Adding Custom Commands

You can add your own commands to the registry:

```go
// Create a custom command
myCommand := commands.Command{
    Name:        "mycommand",
    Description: "Description of my command",
    Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
        // Command implementation
        return nil
    },
}

// Register the command
registry.Register(myCommand)
```

#### Using Progress Callbacks

For long-running operations, use the progress callback system:

```go
progress := commands.NewCommandProgressCallback(bot, userNick, "mycommand")
// Use progress in your operations
progress.OnProgress("Processing...")
```

#### Model Management

You can manage AI models programmatically:

```go
// Get available models
models, exists := faladapter.GetModels("text2image")

// Set current model
err := faladapter.SetCurrentModel("text2image", "fast-sdxl")

// Get current model
model, exists := faladapter.GetCurrentModel("text2image")
```

#### Balance Operations

Access balance operations directly:

```go
// Get user balance
balance, err := dbManager.GetBalance(userID)

// Add balance
err := dbManager.AddBalance(userID, amount)

// Deduct balance
err := dbManager.DeductBalance(userID, amount)
```

### Available Commands

### Basic Commands

#### Help
Shows help information about available commands and models.

**Usage:** `!help [command] [model]`

**Examples:**
- `!help` - Shows general help with current balance and model selections
- `!help text2image` - Shows help for text2image command with available models
- `!help text2image fast-sdxl` - Shows detailed help for the fast-sdxl model

#### Balance
Shows the user's current balance.

**Usage:** `!balance`

#### Rate
Shows current exchange rates for DCR/BTC and DCR/USD.

**Usage:** `!rate`

### Model Configuration

#### ListModels
Lists available models for a specific command type.

**Usage:** `!listmodels <text2image/text2speech/image2image/image2video>`

**Example:** `!listmodels text2image`

#### SetModel
Sets the model to use for a specific command type.

**Usage:** `!setmodel <text2image/text2speech/image2image/image2video> <model_name>`

**Examples:**
- `!setmodel text2image fast-sdxl`
- `!setmodel image2video veo2`

### AI Generation Commands

#### Text2Image
Generates images from text prompts using AI.

**Usage:** `!text2image <prompt>`

**Example:** `!text2image a beautiful sunset over mountains`

#### Image2Image
Transforms images using AI models.

**Usage:** `!image2image <image_url>`

**Example:** `!image2image https://example.com/image.jpg`

Available transformations:
- Ghibli style (ghiblify)
- Cartoon style (cartoonify)
- SVG vectorization (star-vector)

#### Image2Video
Converts images to videos using AI.

**Usage:** `!image2video <image_url> <prompt> [options]`

**Example:** `!image2video https://example.com/image.jpg a cinematic scene --aspect 16:9 --duration 5s`

Available models:
- veo2
- kling-video

**Parameters:**
- `image_url`: URL of the source image
- `prompt`: Description of the desired video animation
- `duration`: Video duration (must be one of: "5s", "6s", "7s", "8s")
- `aspect_ratio`: Aspect ratio (must be one of: "auto", "auto_prefer_portrait", "16:9", "9:16")
- `negative_prompt`: (Optional) Text describing what to avoid (default: blur, distort, and low quality)
- `cfg_scale`: (Optional) Configuration scale (default: 0.5)

**Pricing:**
- veo2: Base price $2.50 for 5 seconds, $0.50 per additional second
- kling-video: Base price $2.0 for 5 seconds, $0.4 per additional second

#### Text2Speech
Converts text to speech using AI.

**Usage:** `!text2speech [voice_id] <text>`

**Example:** `!text2speech Wise_Woman Hello, how are you today?`

Available voices:
- Wise_Woman
- Friendly_Person
- Inspirational_girl
- Deep_Voice_Man
- Calm_Woman
- Casual_Guy
- Lively_Girl
- Patient_Man
- Young_Knight
- Determined_Man
- Lovely_Girl
- Decent_Boy
- Imposing_Manner
- Elegant_Man
- Abbess
- Sweet_Girl_2
- Exuberant_Girl

## Implementation Details

Each command is implemented as a function that returns a `Command` struct:

```go
type Command struct {
    Name        string
    Description string
    Handler     func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error
}
```

The handler function receives:
- `ctx`: Context for the request
- `bot`: The bot instance
- `cfg`: Bot configuration
- `pm`: The received private message
- `args`: Command arguments

## Dependencies

- `github.com/companyzero/bisonrelay/clientrpc/types`: For Bison Relay types
- `github.com/karamble/braibot/internal/faladapter`: For Fal.ai API integration
- `github.com/karamble/braibot/internal/database`: For user balance management
- `github.com/vctt94/bisonbotkit`: For bot functionality
- `github.com/vctt94/bisonbotkit/config`: For configuration

## Configuration

The commands require:
- Fal.ai API key in the bot's configuration under `ExtraConfig["falapikey"]`
- Sufficient user balance for AI operations
- Proper model selection for AI generation commands 