package utils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	kit "github.com/vctt94/bisonbotkit"
)

// SendFileToUser downloads a file from a URL and sends it to a user.
// It creates a temporary file, downloads the content, and sends it using the bot.
// The temporary file is automatically cleaned up after sending.
func SendFileToUser(ctx context.Context, bot *kit.Bot, userNick string, fileURL string, filePrefix string, contentType string) error {
	// Extract file extension from content type
	fileExtension := "bin" // default extension
	if contentType != "" {
		// Split by '/' and take the last part
		parts := strings.Split(contentType, "/")
		if len(parts) > 1 {
			// For types like "image/svg+xml", take "svg+xml"
			fileExtension = parts[len(parts)-1]
			// For types with '+' like "svg+xml", take just "svg"
			if plusIndex := strings.Index(fileExtension, "+"); plusIndex != -1 {
				fileExtension = fileExtension[:plusIndex]
			}
		}
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", filePrefix+"-*."+fileExtension)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temp file when done

	// Download the file
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Copy the data to the temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %v", err)
	}

	// Close the file before sending
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	// Send the file to the user
	if err := bot.SendFile(ctx, userNick, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}

	return nil
}
