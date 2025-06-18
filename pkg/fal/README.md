# Fal Go Client Package (`pkg/fal`)

This Go package provides a client for interacting with the [Fal.ai](https://fal.ai) API, focusing on AI-powered image, video, and speech generation tasks. It's designed to be a reusable, standalone component.

## Features

*   **Asynchronous Task Handling:** Manages the typical Fal.ai workflow (submit -> poll -> get result) for long-running generation tasks.
*   **Model Generation:**
    *   Text-to-Image (`GenerateImage`)
    *   Image-to-Image (`GenerateImage`)
    *   Image-to-Video (`GenerateVideo`)
    *   Text-to-Video (`GenerateVideo`)
    *   Text-to-Speech (`GenerateSpeech`)
*   **Dynamic Model Registration:**
    *   Models are defined in separate files (e.g., `text_image_models.go`).
    *   Models self-register using Go's `init()` mechanism.
    *   Easily add new models by creating a new file implementing the `ModelDefinition` interface.
*   **Model Management:**
    *   Retrieve specific models (`GetModel`).
    *   List available models by type (`GetModels`).
    *   Manage the currently active default model per type (`GetCurrentModel`, `SetCurrentModel`).
*   **Progress Tracking:** Provides a `ProgressCallback` interface for monitoring task status (queue position, logs, progress updates, errors).
*   **Extensible:** Designed with interfaces and clear separation for potential future enhancements.

## Installation

```bash
# Assuming this package is part of a larger project
# Go modules will handle vendoring/dependencies
go get <your-project-module-path>/pkg/fal 
```
*(Adjust the module path as needed)*

## Usage

### 1. Creating a Client

```go
import "<your-project-module-path>/pkg/fal"

// Create a new client with your API key
client := fal.NewClient("your-fal-api-key")

// Optionally enable debug logging
clientWithDebug := fal.NewClient("your-fal-api-key", fal.WithDebug(true))

// Optionally provide a custom HTTP client
customHTTPClient := &http.Client{ Timeout: 60 * time.Second }
clientWithCustomHTTP := fal.NewClient("your-fal-api-key", fal.WithHTTPClient(customHTTPClient))
```

### 2. Defining a Progress Callback (Optional)

Implement the `fal.ProgressCallback` interface to receive updates.

```go
type MyProgressTracker struct {
    // ... your fields (e.g., logger, channel)
}

func (t *MyProgressTracker) OnQueueUpdate(position int, eta time.Duration) {
    fmt.Printf("Queue Position: %d, ETA: %v\n", position, eta)
}
func (t *MyProgressTracker) OnLogMessage(message string) {
    fmt.Printf("Log: %s\n", message)
}
func (t *MyProgressTracker) OnProgress(status string) {
    fmt.Printf("Status: %s\n", status)
}
func (t *MyProgressTracker) OnError(err error) {
    fmt.Printf("Error: %v\n", err)
}

// Instantiate your callback
var progressCallback fal.ProgressCallback = &MyProgressTracker{} 
```

### 3. Performing Generation Tasks

**Text-to-Image (e.g., Flux Schnell):**

```go
// Use the specific request struct for type safety and defined options
req := fal.FluxSchnellRequest{
	BaseImageRequest: fal.BaseImageRequest{
		Prompt:   "a hyperrealistic cat wearing sunglasses",
		Progress: progressCallback, // Optional
	},
	// Set specific options for this model
	NumImages: 2,
	ImageSize: "square",
}

// Pass the specific request struct to GenerateImage
resp, err := client.GenerateImage(context.Background(), &req)
if err == nil {
	fmt.Println("Image URL:", resp.Images[0].URL)
	// Access seed if needed: fmt.Println("Seed:", resp.Seed)
}
```

**Image-to-Image (e.g., Ghiblify):**

```go
// Use the specific request struct
req := fal.GhiblifyRequest{
	BaseImageRequest: fal.BaseImageRequest{
		ImageURL: "https://example.com/input.jpg",
		Progress: progressCallback, // Optional
		// Prompt is optional for ghiblify but can be added here
	},
	// Ghiblify currently has no specific options beyond BaseImageRequest
}

// Pass the specific request struct to GenerateImage
resp, err := client.GenerateImage(context.Background(), &req)
// ... handle response ...
```

**Text-to-Speech (e.g., Minimax TTS):**

```go
// Use the specific request struct
req := fal.MinimaxTTSRequest{
	BaseSpeechRequest: fal.BaseSpeechRequest{
		Text:     "Hello from the digital world!",
		Progress: progressCallback, // Optional
	},
	// Set specific options for this model
	VoiceID: "Calm_Woman",
	Speed:   floatPtr(0.9), // Use helper for optional pointers if needed
}

// Pass the specific request struct to GenerateSpeech
resp, err := client.GenerateSpeech(context.Background(), &req)
if err == nil {
    fmt.Println("Audio URL:", resp.AudioURL)
}

// Helper to get pointer for optional float (example)
func floatPtr(f float64) *float64 { return &f }
```

**Image-to-Video (e.g., Veo2):**

```go
// Use the specific request struct for type safety and defaults
req := fal.Veo2Request{
	BaseVideoRequest: fal.BaseVideoRequest{
		Prompt:   "make the cat dance",
		ImageURL: "https://example.com/cat.jpg",
		Progress: progressCallback, // Attach progress here
	},
	// Specific Veo2 options (will use defaults defined in model if empty)
	Duration:    "8",
	AspectRatio: "16:9",
}

// Pass the specific request struct to GenerateVideo
resp, err := client.GenerateVideo(context.Background(), &req)
// ... handle response ...
```

**Text-to-Video (e.g., Kling Text):**

```go
// Use the specific request struct (KlingVideoRequest handles text/image variants)
req := fal.KlingVideoRequest{
	BaseVideoRequest: fal.BaseVideoRequest{
		Prompt:   "a teddy bear playing drums on the moon",
		Progress: progressCallback,
	},
	// Specific Kling options
	Duration:    "10",
	AspectRatio: "16:9",
}

// Pass the specific request struct to GenerateVideo
resp, err := client.GenerateVideo(context.Background(), &req)
// ... handle response ...
```

### 4. Managing Models

```go
// Get a specific model's details
model, exists := fal.GetModel("fast-sdxl", "text2image")
if exists {
    fmt.Printf("Model: %s, Price: $%.2f\n", model.Name, model.PriceUSD)
}

// List all text-to-image models
t2iModels, _ := fal.GetModels("text2image")
for name, model := range t2iModels {
    fmt.Println("- ", name, model.Description)
}

// Get the currently set default model for text-to-image
currentModel, exists := fal.GetCurrentModel("text2image")
if exists {
   fmt.Println("Current default t2i model:", currentModel.Name) 
}

// Set a new default model for text-to-image
err := fal.SetCurrentModel("text2image", "hidream-i1-full")
if err != nil {
    fmt.Println("Error setting default model:", err)
}
```

## Adding New Models

1.  Create a new Go file in the `pkg/fal` directory (e.g., `my_new_model.go`).
2.  Define a struct for your model (it can be empty: `type myNewModelDefinition struct{}`).
3.  Implement the `fal.ModelDefinition` interface for your struct:
    ```go
    func (m *myNewModelDefinition) Define() fal.Model {
        return fal.Model{
            Name:        "your-model-id", // The ID fal.ai uses
            Description: "Description of your model",
            PriceUSD:    0.05, // Cost per run
            Type:        "text2image", // or image2image, text2speech, etc.
            HelpDoc:     "Usage instructions...",
            Options:     &YourModelOptions{}, // Optional: Define and link options struct
        }
    }
    ```
4.  In the same file, add an `init()` function to register your model definition:
    ```go
    func init() {
        fal.registerModel(&myNewModelDefinition{})
    }
    ```
5.  If your model requires specific options, define a struct for them (e.g., `YourModelOptions`) and implement `GetDefaultValues()` and `Validate()` methods for your struct. These methods are called by the framework for defaulting and validation.

The model will now be available via `GetModel`, `GetModels`, and can be used in generation requests by its `Name`.

## Error Handling

API errors are typically returned as `*fal.Error`:

```go
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *Error) Error() string {
    return e.Message
}
```
Check for this type to handle API-specific issues gracefully. Other standard Go errors may be returned for network issues, decoding problems, etc.

## License

This package is licensed under the ISC License - see the [LICENSE](../../LICENSE) file for details.