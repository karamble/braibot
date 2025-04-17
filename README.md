# Braibot

**Braibot** is a chatbot for the [Bison Relay](https://bisonrelay.org/) platform, you can find [Bison Relay on Github](https://github.com/companyzero/bisonrelay/). It is using the [fal.ai API](https://fal.ai/) to generate AI-powered images and speech from text prompts. Payments are processed via microtransactions on [Decred's lightning network](https://docs.decred.org/lightning-network/overview/).

This bot is based on the [BisonBotKit](https://github.com/vctt94/bisonbotkit) by @vctt94

## Features

- **AI Content Generation**: 
  - Create images using various AI models via fal.ai
  - Generate speech from text using AI models
  - Customizable model selection for both image and speech generation
- **Lightning Network Payments**: 
  - Users can send tips to the Bot on Bison Relay using Decred's lightning network
  - Balances are stored in a SQLite3 database
  - Real-time balance tracking and updates
- **Bison Relay Integration**: 
  - Seamless private messaging interface
  - Command-based interaction system
  - Welcome messages for new users
- **Model Management**: 
  - List available AI models
  - Check current rates
  - Switch between different models
- **Help System**:
  - Comprehensive command documentation
  - Context-aware help for specific commands
  - Model-specific help information
  - Current balance and model selection status
- **Debug Mode**: Built-in debugging capabilities for development and troubleshooting

## Project Structure

```
braibot/
├── internal/
│   ├── audio/         # Audio processing and generation
│   ├── commands/      # Bot command implementations
│   ├── config/        # Configuration management
│   ├── database/      # Database operations
│   ├── faladapter/    # fal.ai API integration
│   └── utils/         # Utility functions
├── pkg/
│   └── fal/           # fal.ai API package
├── main.go            # Main application entry point
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
└── README.md          # Project documentation
```

## Prerequisites

- **Go**: [Download Go](https://go.dev/dl/) (version 1.16 or higher recommended)
- **Bison Relay Account**: Active account and installation of Bison Relay
- **fal.ai API Key**: Get one at [fal.ai](https://fal.ai/) and fund your account balance for AI usage
- **Decred**: Find out where to get Decred on [Decred's Exchange website](https://decred.org/exchanges/)
- **SQLite3**: Required for database operations

## Installation

1. **Clone Repository**  
   ```bash
   git clone https://github.com/karamble/braibot.git
   ```

2. **Enter Directory**  
   ```bash
   cd braibot
   ```

3. **Install Dependencies**
   ```bash
   go mod tidy
   ```

4. **Build Project**
   ```bash
   go build
   ```

## Configuration

1. **Bison Relay Configuration**
   Edit your Bison Relay configuration to activate the [clientrpc] functions. Set the following parameters:
   ```
   jsonrpclisten=<your-rpc-listen-address>
   rpccertpath=<path-to-rpc-cert>
   rpckeypath=<path-to-rpc-key>
   rpcclientcapath=<path-to-client-ca>
   rpcissueclientcert=1
   ```

2. **Bot Configuration**
   On first launch, the bot automatically:
   - Looks for the BisonRelay configuration
   - Creates its configuration file in `~/.braibot/braibot.conf`
   - Sets up necessary directories and database
   - Will ask for the fal-api key if not present

   Edit `~/.braibot/braibot.conf` and add:
   ```
   falapikey=<your-fal-ai-api-key>
   ```

## Running the Bot

1. **Start Bison Relay Client**
   Make sure your Bison Relay client is running first.

2. **Launch the Bot**
   ```bash
   # Option 1: Run directly
   go run .

   # Option 2: Run built binary
   ./braibot

   # Option 3: Run with custom app root
   ./braibot -approot /path/to/custom/dir
   ```

## Using the Bot

Send private messages on Bison Relay to the Bot:

- **`!rate`**  
  Shows exchange rates for DCR/BTC and DCR/USD provided by Bison Relay

- **`!listmodels <text2image/text2speech/image2video/image2image>`**  
  Lists available AI models.

- **`!setmodel <text2image/text2speech/image2video/image2image> <model_name>`**  
  Sets model to use for any of the supported AI commands (use model from `!listmodels`).  
  Examples: 
  - `!setmodel text2image fast-sdxl`
  - `!setmodel image2video stable-video-diffusion`
  - `!setmodel image2image stable-diffusion`

- **`!text2image <prompt>`**  
  Generates an image with the AI model previously specified with the setmodel command.  
  Example: `!text2image a starry night`

- **`!text2speech <prompt>`**  
  Generates an AI audio clip from text prompt.  
  Example: `!text2speech "Hello world"`

- **`!help`**  
  Lists available commands.

- **`!balance`**
  Users can query their current debit balance

## Troubleshooting

1. **Bot Won't Start**
   - Ensure Bison Relay client is running
   - Check configuration files exist and are properly formatted
   - Verify fal.ai API key is valid and has sufficient balance

2. **Commands Not Working**
   - Check if you have sufficient balance
   - Verify the selected model is available
   - Ensure proper command syntax

3. **Database Issues**
   - Check permissions on the app root directory
   - Verify SQLite3 is installed
   - Check log files for specific errors

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the ISC License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bison Relay](https://github.com/companyzero/bisonrelay/) for the messaging platform
- [BisonBotKit](https://github.com/vctt94/bisonbotkit) for the bot framework
- [fal.ai](https://fal.ai/) for AI capabilities
- [Decred](https://decred.org/) for the Lightning Network implementation
