# Fal.ai API Client

A standalone Go client for interacting with the [Fal.ai](https://fal.ai) API, providing AI-powered image, video, and speech generation capabilities.

## Features

- **Text-to-Image Generation**
  - Multiple model options with different quality levels
  - Customizable parameters
  - Progress tracking
  - Queue management

- **Image-to-Image Transformation**
  - Ghibli style conversion
  - Cartoon style conversion
  - SVG vectorization
  - Customizable parameters

- **Image-to-Video Conversion**
  - Multiple video models (Veo2, Kling-video)
  - Customizable duration and aspect ratio
  - Progress tracking
  - Queue management

- **Text-to-Speech Generation**
  - Multiple voice options
  - Customizable parameters
  - Progress tracking

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

### Available Models

#### Text-to-Image Models
- `fast-sdxl` - Fast model for quick image generation
- `hidream-i1-full` - High-quality model for detailed images
- `hidream-i1-dev` - Development version of HiDream
- `hidream-i1-fast` - Faster version of HiDream
- `flux-pro/v1.1` - Professional model
- `flux-pro/v1.1-ultra` - Ultra version of professional model
- `flux/schnell` - Quick model for rapid generation

#### Image-to-Image Models
- `ghiblify` - Transforms images into Studio Ghibli style
- `cartoonify` - Transforms images into Pixar-like 3D cartoon style
- `star-vector` - Converts images to SVG using AI vectorization

#### Text-to-Speech Models
- `minimax-tts/text-to-speech`
  - Multiple voice options available
  - High-quality speech synthesis

#### Image-to-Video Models
- `veo2`
  - Creates videos from images with realistic motion
  - Supports multiple aspect ratios
  - Duration options: 5-8 seconds
- `kling-video`
  - Advanced video generation
  - Customizable parameters
  - Negative prompt support

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

// Create a video request with Veo2
req := fal.Veo2Request{
    BaseVideoRequest: fal.BaseVideoRequest{
        Prompt:   "a cinematic scene",
        ImageURL:  "https://example.com/image.jpg",
        Model:     "veo2",
        Progress:  progress,
    },
    Duration:    "5",
    AspectRatio: "16:9",
}

// Generate the video
resp, err := client.GenerateVideo(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Access the generated video
videoURL := resp.GetURL()
```

### Generating Speech from Text

```go
// Create a progress callback
progress := &MyProgressCallback{}

// Create a speech request
req := fal.SpeechRequest{
    Text:     "Hello, how are you today?",
    VoiceID:  "Wise_Woman",
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

## Progress Tracking

The client supports progress tracking through the `ProgressCallback` interface:

```go
type ProgressCallback interface {
    OnQueueUpdate(position int, eta time.Duration)
    OnLogMessage(message string)
    OnProgress(status string)
    OnError(err error)
}
```

## Error Handling

The client provides detailed error information through the `Error` type:

```go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

## License

This package is licensed under the ISC License - see the [LICENSE](../../LICENSE) file for details. 