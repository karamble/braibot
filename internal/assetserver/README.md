# Asset Server Client Package

This package provides a client for interacting with the braibot-assetserver, which handles file uploads for the braibot application.

## Features

- Simple file upload interface
- Automatic multipart form handling
- Error handling and response parsing
- Configurable server URL and API key

## Usage

```go
import "github.com/karamble/braibot/internal/assetserver"

// Create a new client
client := assetserver.NewClient("https://assets.example.com", "your-api-key")

// Upload a file
fileData := []byte("file contents")
resp, err := client.UploadFile("example.jpg", fileData)
if err != nil {
    // Handle error
}

// Use the download URL
downloadURL := resp.URL
```

## Configuration

The client requires two configuration parameters:
- `serverURL`: The base URL of the assetserver (e.g., "https://assets.example.com")
- `apiKey`: The API key for authentication

## Response Format

The upload response includes:
- `Success`: Boolean indicating if the upload was successful
- `Message`: Status message from the server
- `URL`: Download URL for the uploaded file (only present on success)

## Error Handling

The client provides detailed error messages for:
- Network errors
- Authentication failures
- Server errors
- File upload failures 