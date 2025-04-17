# Fal.ai API Client

A standalone Go client for interacting with the [Fal.ai](https://fal.ai) API, providing AI-powered image and video generation capabilities.

## Features

- Text-to-Image generation with multiple model options
- Image-to-Image transformation (Ghibli style and cartoon style)
- Image-to-Video conversion
- Text-to-Speech generation with voice selection
- Queue management and progress tracking
- Model management and configuration
- Debug mode for troubleshooting

## Installation

```bash
go get github.com/karamble/braibot/pkg/fal
```

## Usage

### Creating a Client

```go
import "github.com/karamble/braibot/pkg/fal"

// Create a new client with default options
client := fal.NewClient("your-api-key")

// Create a client with debug mode
client := fal.NewClient("your-api-key", fal.WithDebug(true))

// Create a client with custom HTTP client
client := fal.NewClient("your-api-key", fal.WithHTTPClient(&http.Client{
    Timeout: 60 * time.Second,
}))
```

### Generating Images from Text

```go
// Create a progress callback
progress := &MyProgressCallback{}

// Create an image request
req := fal.ImageRequest{
    Prompt:   "a beautiful sunset over mountains",
    Model:    "fast-sdxl",
    Options:  map[string]interface{}{"num_images": 1},
    Progress: progress,
}

// Generate the image
resp, err := client.GenerateImage(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Access the generated image
imageURL := resp.Images[0].URL
```

### Transforming Images

```go
// Create a progress callback
progress := &MyProgressCallback{}

// Create an image request
req := fal.ImageRequest{
    Model:    "ghiblify",
    Options:  map[string]interface{}{
        "image_url": "https://example.com/image.jpg",
    },
    Progress: progress,
}

// Transform the image
resp, err := client.GenerateImage(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Access the transformed image
imageURL := resp.Images[0].URL
```

### Converting Images to Video

```go
// Create a progress callback
progress := &MyProgressCallback{}

// Create a video request
req := fal.KlingVideoRequest{
    BaseVideoRequest: fal.BaseVideoRequest{
        Prompt:   "a cinematic scene",
        ImageURL:  "https://example.com/image.jpg",
        Model:     "kling-video",
        Progress:  progress,
        Options:   make(map[string]interface{}),
    },
    Duration:       "5",
    AspectRatio:    "16:9",
    NegativePrompt: "blur, distort, and low quality",
    CFGScale:       0.5,
}

// Generate the video
resp, err := client.GenerateVideo(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Access the generated video
videoURL := resp.Video.URL
```

### Generating Speech

```go
// Create a progress callback
progress := &MyProgressCallback{}

// Create a speech request
req := fal.SpeechRequest{
    Text:     "Hello, world!",
    VoiceID:  "minimax-tts/text-to-speech",
    Options:  map[string]interface{}{"speed": 1.0},
    Progress: progress,
}

// Generate the speech
resp, err := client.GenerateSpeech(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Access the generated audio
audioURL := resp.AudioURL
```

### Progress Tracking

Implement the `ProgressCallback` interface to track progress:

```go
type MyProgressCallback struct{}

func (c *MyProgressCallback) OnQueueUpdate(position int, estimatedTime time.Duration) {
    fmt.Printf("Queue position: %d, ETA: %v\n", position, estimatedTime)
}

func (c *MyProgressCallback) OnProgress(percentage int, status string) {
    fmt.Printf("Progress: %d%%, Status: %s\n", percentage, status)
}

func (c *MyProgressCallback) OnError(err error) {
    fmt.Printf("Error: %v\n", err)
}

func (c *MyProgressCallback) OnLogMessage(message string) {
    fmt.Printf("Log: %s\n", message)
}
```

### Model Management

```go
// Get available models
models, exists := fal.GetModels("text2image")
if !exists {
    log.Fatal("text2image models not found")
}

// Get current model
model, exists := fal.GetCurrentModel("text2image")
if !exists {
    log.Fatal("current model not found")
}

// Set current model
err := fal.SetCurrentModel("text2image", "fast-sdxl")
if err != nil {
    log.Fatal(err)
}
```

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

### Image-to-Image Models

| Model Name | Description | Price (USD) |
|------------|-------------|-------------|
| ghiblify | Transforms images into Studio Ghibli style artwork | $0.02 |
| cartoonify | Transforms images into cartoon style artwork | $0.02 |

### Image-to-Video Models

| Model Name | Description | Price (USD) |
|------------|-------------|-------------|
| kling-video/v2 | Convert images to videos with motion | $2.00 base + $0.40 per additional second |

### Text-to-Speech Models

| Model Name | Description | Price |
|------------|-------------|--------|
| minimax-tts/text-to-speech | Text-to-speech model with multiple voices | $0.10 per 1000 characters |

Available Voices:
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

## Error Handling

The package provides detailed error information through the `Error` type:

```go
type Error struct {
    Code    string
    Message string
}
```

Common error codes:
- `INVALID_MODEL`: The specified model does not exist
- `GENERATION_FAILED`: The generation process failed
- `NO_IMAGES`: No images were generated
- `NO_AUDIO`: No audio was generated
- `NO_VIDEO`: No video was generated
- `INVALID_REQUEST`: The request parameters are invalid

## Thread Safety

The client is designed to be thread-safe and can handle concurrent requests. All methods are safe to call from multiple goroutines.

## License

ISC License 