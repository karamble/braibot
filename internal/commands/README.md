# Internal Commands Package (`internal/commands`)

This package defines and registers the chat commands available within Braibot. It translates user input (like `!help` or `!text2image`) into actions performed by the bot, often interacting with other internal packages (`database`, `faladapter`, `image`, `video`, `audio`, `utils`) and the underlying `pkg/fal` library.

## Core Components

*   **Command Struct:** Defines the structure for a command (Name, Description, Handler).
*   **Registry:** A central registry (`Registry`) to hold and retrieve all registered `Command` structs.
*   **Initialization:** The `InitializeCommands` function creates the registry and registers all standard Braibot commands.
*   **Command Parsing:** The `IsCommand` function checks if an incoming message is a valid command and extracts the command name and arguments.
*   **Command Handlers:** Individual files (e.g., `text2image.go`, `help.go`, `balance.go`) contain the specific logic executed when a command is invoked.

## Available Commands

This package implements the handlers for the following user-facing commands:

### Basic & Informational

*   **`!help`**: Shows general or command/model-specific help.
*   **`!balance`**: Displays the user's current DCR balance.
*   **`!rate`**: Shows current DCR exchange rates.

### Model Management

*   **`!listmodels [task]`**: Lists available AI models for a task (`text2image`, `image2image`, `text2speech`, `image2video`, `text2video`).
*   **`!setmodel [task] [model_name]`**: Sets the default AI model for a specific task.

### AI Generation

*   **`!text2image [prompt]`**: Generates an image from text.
*   **`!image2image [image_url] [optional prompt]`**: Transforms an image using AI.
*   **`!image2video [image_url] [optional prompt]`**: Generates a video from an image.
*   **`!text2video [prompt]`**: Generates a video from text.
*   **`!text2speech [optional_voice_id] [text]`**: Generates speech audio from text.

Refer to the specific handler files (e.g., `text2image.go`) for the detailed implementation logic of each command. Interactions with external APIs (Fal.ai) are typically mediated through the `internal/faladapter` package, and balance checks utilize the `internal/database` and `internal/utils` packages.

## Features

- **Debug Mode**: Built-in debugging capabilities for development and troubleshooting

## Integration Guide

### Package Structure

The commands package is part of a larger system with the following structure:

```
braibot/
├── pkg/
│   └── fal/           # Core Fal.ai API client
├── internal/
│   ├── audio/         # Audio processing and generation
│   ├── commands/      # Command handlers
│   ├── config/        # Configuration management
│   ├── database/      # Database operations
│   ├── faladapter/    # Fal.ai API integration
│   ├── image/         # Image processing
│   ├── video/         # Video processing
│   └── utils/         # Utility functions
```

### Prerequisites

1. **BisonBotKit Setup**
   - A working BisonBotKit project
   - Bison Relay client running with RPC enabled
   - Proper configuration for BisonBotKit

2. **Required Dependencies**
   - Fal.ai API key
   - A Bison Relay Client running with json rpc interface enabled
   - Go 1.16 or higher
   - SQLite3

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
       LogFile:        filepath.Join(appRoot, "logs", "braibot.log"),
       DebugLevel:     "info",
       MaxLogFiles:    5,
       MaxBufferLines: 1000,
   })
   if err != nil {
       log.Fatal(err)
   }
   defer logBackend.Close()

   // Load bot configuration
   cfg, err := botkitconfig.LoadBotConfig(appRoot, "braibot.conf")
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
   // - AI commands: text2image, text2speech, image2image, image2video, text2video
   commandRegistry := commands.InitializeCommands(dbManager, debug)
   ```

4. **Set Up Message Handler**
   ```go
   // Create channels for receiving messages and tips
   pmChan := make(chan types.ReceivedPM)
   tipChan := make(chan types.ReceivedTip)
   tipProgressChan := make(chan types.TipProgressEvent)

   // Set up PM channels/log
   cfg.PMChan = pmChan
   cfg.PMLog = logBackend.Logger("PM")

   // Set up tip channels/logs
   cfg.TipLog = logBackend.Logger("TIP")
   cfg.TipProgressChan = tipProgressChan
   cfg.TipReceivedLog = logBackend.Logger("TIP_RECEIVED")
   cfg.TipReceivedChan = tipChan

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

The bot requires the following configuration:

1. **Bison Relay Configuration**
   Edit your Bison Relay configuration to activate the [clientrpc] functions:
   ```
   jsonrpclisten=<your-rpc-listen-address>
   rpccertpath=<path-to-rpc-cert>
   rpckeypath=<path-to-rpc-key>
   rpcclientcapath=<path-to-client-ca>
   rpcissueclientcert=1
   ```

2. **Bot Configuration**
   The bot automatically creates its configuration file in `~/.braibot/braibot.conf` and will prompt for the fal.ai API key if not present. Add:
   ```
   falapikey=<your-fal-ai-api-key>
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

## Project Structure

```
braibot/
├── pkg/
│   └── fal/           # Core Fal.ai API client
├── internal/
│   ├── audio/         # Audio processing and generation
│   ├── commands/      # Command handlers
│   ├── config/        # Configuration management
│   ├── database/      # Database operations
│   ├── faladapter/    # Fal.ai API integration
│   ├── image/         # Image processing
│   ├── video/         # Video processing
│   └── utils/         # Utility functions
``` 