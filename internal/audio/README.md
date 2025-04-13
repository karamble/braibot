Ogg encoding originally from https://github.com/chenbh/skynetbot.

# Audio Package for BisonBotKit

This package provides audio conversion functionality for [BisonBotKit](https://github.com/vctt94/bisonbotkit), specifically designed to convert PCM audio data to Opus format with OGG container for use in BisonRelay chat messages as voice messages.

## Overview

The audio package enables the conversion of raw PCM audio data (typically at 24kHz sample rate) to Opus-encoded audio in an OGG container format. This format is compatible with the BisonRelay client's internal audio player, allowing users to send and receive voice messages in chat.

## Features

- PCM to Opus conversion with optimal settings for voice messages
- OGG container formatting for compatibility with BisonRelay
- Configurable audio parameters (sample rate, channels, bitrate)
- Efficient encoding for voice communication
- Thread-safe implementation

## Usage

### Converting PCM to Opus

```go
import "github.com/karamble/braibot/internal/audio"

// Convert PCM data to Opus format with OGG container
opusData, err := audio.ConvertPCMToOpus(pcmData)
if err != nil {
    // Handle error
}

// Use the opusData in your BisonRelay message
// For example, encode as base64 and include in an embed
encodedAudio := base64.StdEncoding.EncodeToString(opusData)
message := fmt.Sprintf("--embed[alt=Voice Message,type=audio/ogg,data=%s]--", encodedAudio)
```

## Audio Parameters

The package uses the following default parameters for optimal voice quality:

- Sample Rate: 24000 Hz (24 kHz)
- Channels: 1 (Mono)
- Bitrate: 64000 bps (64 kbps)
- Frame Size: 960 samples (40ms at 24kHz)

These parameters are optimized for voice communication and provide a good balance between quality and file size.

## Integration with BisonBotKit

This package is designed to work seamlessly with BisonBotKit, allowing you to:

1. Convert audio from text-to-speech services to a format compatible with BisonRelay
2. Send voice messages in chat using the BisonRelay embed format
3. Process audio data from various sources for use in your bot

## Dependencies

- [github.com/companyzero/gopus](https://github.com/companyzero/gopus) - Go bindings for the Opus audio codec

## Implementation Details

The package consists of several components:

- `ConvertPCMToOpus`: Main function for converting PCM data to Opus format
- `OpusWriter`: Handles writing Opus packets to the output stream
- `OggWriter`: Manages the OGG container format

## Example Workflow

1. Generate audio data from a text-to-speech service
2. Convert the raw PCM data to Opus format with OGG container
3. Encode the resulting data as base64
4. Include the encoded data in a BisonRelay message using the embed format
5. Send the message to the user

