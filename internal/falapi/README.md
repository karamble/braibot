# Fal.ai API Client Package for BisonBotKit

This package provides a Go client for interacting with the [Fal.ai](https://fal.ai) API, specifically designed for text-to-image and text-to-speech generation services. It's built to handle asynchronous API requests with queue management and progress tracking. The package is designed to work with [BisonBotKit](https://github.com/vctt94/bisonbotkit)

## Features

- Text-to-Image generation with multiple model options
- Text-to-Speech generation with voice selection
- Queue management and status tracking
- Progress updates through bot messaging
- Model management and configuration
- Debug mode for troubleshooting

## Available Models

### Text-to-Image Models

| Model Name | Description | Price (USD) |
|------------|-------------|-------------|
| fast-sdxl | Fast model for generating images quickly | $0.02 |
| hidream-i1-full | High-quality model for detailed images | $0.10 |
| hidream-i1-dev | Development version of the HiDream model | $0.06 |
| hidream-i1-fast | Faster version of the HiDream model | $0.03 |
| flux-pro/v1.1 | Professional model for high-end image generation | $0.08 |
| flux-pro/v1.1-ultra | Ultra version of the professional model | $0.12 |
| flux/schnell | Quick model for rapid image generation | $0.02 |

### Text-to-Speech Models

| Model Name | Description | Price |
|------------|-------------|--------|
| minimax-tts/text-to-speech | Text-to-speech model for converting text to audio | $0.10 per 1000 characters |

## Usage

### Creating a Client

```go
client := falapi.NewClient(apiKey, debug)
```

### Generating Images

```go
imageResp, err := client.GenerateImage(ctx, prompt, modelName, bot, userNick)
if err != nil {
    // Handle error
}
```

### Generating Speech

```go
audioResp, err := client.GenerateSpeech(ctx, text, voiceID, bot, userNick)
if err != nil {
    // Handle error
}
```

### Managing Models

```go
// Get available models for a command type
models, err := client.GetModels("text2image")

// Get current model for a command type
model, err := client.GetCurrentModel("text2image")

// Set current model for a command type
err := client.SetCurrentModel("text2image", "fast-sdxl")
```

## Response Types

### ImageResponse
Contains information about generated images including:
- Image URLs
- Dimensions
- Content types
- Generation timings
- NSFW detection
- Original prompt

### AudioResponse
Contains information about generated audio including:
- Audio URL
- Content type
- File name
- File size
- Duration

## Queue Management

The package automatically handles queue management for API requests:
- Tracks queue position
- Provides status updates
- Supports progress monitoring
- Handles request cancellation

## Error Handling

The package provides detailed error information for:
- API request failures
- Queue management issues
- Invalid model selections
- Generation failures

## Debug Mode

Enable debug mode when creating the client to get detailed logging information:
```go
client := falapi.NewClient(apiKey, true)
```

## Dependencies

- Standard Go libraries only
- No external dependencies required

## Thread Safety

The client is designed to be thread-safe and can handle concurrent requests.

## Note

This package is designed to work with a bot interface that implements a `SendPM` method for progress updates. The bot interface should have the following signature:

```go
SendPM(ctx context.Context, nick string, message string) error
``` 