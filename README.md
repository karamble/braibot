# Braibot

**Braibot** is a chatbot for the [Bison Relay](https://bisonrelay.org/) platform, you can find [Bison Relay on Github](https://github.com/companyzero/bisonrelay/). It is using the [fal.ai API](https://fal.ai/) to generate AI-powered images and speech from text prompts. Payments are processed via microtransactions on [Decred's lightning network](https://docs.decred.org/lightning-network/overview/).

This bot is based on the [BisonBotKit](https://github.com/vctt94/bisonbotkit) by @vctt94

## Features

- **AI Content Generation**: Create images and speech with AI models.
- **Lightning Network Payments**: Users can send tips to the Bot on Bison Relay using Decred's lightning network. Balances are stored in a sqllite3 database.
- **Bison Relay Integration**: Send commands via private messages.
- **Model Management**: List and select AI models, check rates.

## Prerequisites

- **Go**: [Download Go](https://go.dev/dl/).
- **Bison Relay Account**: Active account and installation of Bison Relay.
- **fal.ai API Key**: Get one at [fal.ai](https://fal.ai/) and fund your account balance for ai usage.
- **Decred **: Find out where to get Decred on [Decred's Exchange website](https://decred.org/exchanges/).

## Installation

1. **Clone Repository**  
   ```
   git clone https://github.com/karamble/braibot.git
   ```
2. **Enter Directory**  
   ```
   cd braibot
   ```
3. **Install Dependencies**
   ```
   go mod tidy
   ```
4. **Build Project**
   ```
   go build
   ```

## Configuration

Edit your Bison Relay configuration to activate the [clientrpc] functions. Parameters jsonrpclisten, rpccertpath, rpckeypath, rpcclientcapath and rpcissueclientcert should be set to enable the RPC interface.

On first launch, the bot automatically looks for the BisonRelay configuration and creates its own configuration file in its appdata dir (~/.braibot/braibot.conf). Once the configuration file is created you have to manually edit it and add your fal.api key with the parameter falapikey=

## Running the Bot

Make sure you first launch your Bison Relay Client. Then launch the Bot, it will connect to your Bison Relay Client via RPC. If you do not want to go build or go install the bot you can simply run it with the command
```
go run .
```

## Using the Bot

Send private messages on Bison Relay to the Bot:

- **`!rate`**  
  Shows exchange rates for DCR/BTC and DCR/USD provided by Bison Relay

- **`!listmodels <text2image/text2speech>`**  
  Lists available AI models.

- **`!setmodel <text2image/text2speech> <model_name>`**  
  Sets model to use (use model from `!listmodels`).  
  Example: `!setmodel text2image fast-sdxl`

- **`!text2image <prompt>`**  
  Generates an image with the ai model previously specified with the setmodel command.  
  Example: `!text2image a starry night`

- **`!text2speech <prompt>`**  
  Generates an ai audio clip from text prompt.  
  Example: `!text2speech "Hello world"`

- **`!help`**  
  Lists available commands.

- **`!balance`**
  Users can query their current debit balance

**Note**: Make sure to have sufficient inbound liquidity to receive funds from users. Trough the Bison Relay 'Send Tip' function users can deposit funds for ai usage. Their current balance is stored in a sqllite3 database.
