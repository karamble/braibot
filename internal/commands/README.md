# Commands Package

This package provides command handlers for the Braibot, a chatbot for the Bison Relay platform. The commands package implements various AI-powered features using the Fal.ai API.

## Overview

The commands package contains handlers for different bot commands that users can invoke through private messages. Each command is implemented as a function that returns a `Command` struct with a name, description, and handler function.

## Available Commands

### Text2Image

Generates images from text prompts using AI.

**Usage:** `!text2image [prompt]`

**Example:** `!text2image a beautiful sunset over mountains`

This command uses the Fal.ai API to generate images based on the provided text prompt. The generated image is sent back to the user as an embedded image in a private message.

### Text2Speech

Converts text to speech using AI.

**Usage:** `!text2speech [voice_id] [text]`

**Example:** `!text2speech Wise_Woman Hello, how are you today?`

**Available Voices:**
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

This command uses the Fal.ai API to convert text to speech using the specified voice. The generated audio is sent back to the user as an embedded audio file in a private message.

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
- `github.com/karamble/braibot/internal/falapi`: For Fal.ai API integration
- `github.com/vctt94/bisonbotkit`: For bot functionality
- `github.com/vctt94/bisonbotkit/config`: For configuration

## Configuration

The commands require a Fal.ai API key to be set in the bot's configuration under `ExtraConfig["falapikey"]`. 