package speech

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	// "github.com/companyzero/bisonrelay/clientrpc/types" // Only needed for old billing call
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// SpeechService handles speech generation
type SpeechService struct {
	client         *fal.Client
	dbManager      *database.DBManager
	bot            *kit.Bot
	debug          bool
	billingEnabled bool // Added billing enabled flag
}

// NewSpeechService creates a new SpeechService
func NewSpeechService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool, billingEnabled bool) *SpeechService {
	return &SpeechService{
		client:         client,
		dbManager:      dbManager,
		bot:            bot,
		debug:          debug,
		billingEnabled: billingEnabled, // Store the flag
	}
}

// GenerateSpeech generates speech based on the internal request, handling billing conditionally.
func (s *SpeechService) GenerateSpeech(ctx context.Context, req *SpeechRequest) (*SpeechResult, error) {
	// 1. Calculate cost and CHECK balance if billing is enabled
	var requiredDCR, currentBalanceDCR float64
	var checkErr error
	if s.billingEnabled {
		// Call CheckBalance, which now returns the error directly if insufficient or other issue
		requiredDCR, currentBalanceDCR, checkErr = utils.CheckBalance(ctx, s.dbManager, req.UserID[:], req.PriceUSD, s.debug, s.billingEnabled)
		if checkErr != nil {
			// Return the error (could be ErrInsufficientBalance or another critical error)
			// The calling layer (main.go) will handle ErrInsufficientBalance specifically.
			return &SpeechResult{Success: false, Error: checkErr}, checkErr
		}
	}

	// 2. Send initial message (adjusted for billing status)
	var infoMsg string
	if s.billingEnabled {
		infoMsg = fmt.Sprintf("Request cost: $%.2f USD (%.8f DCR). Your balance: %.8f DCR. Processing speech request...", req.PriceUSD, requiredDCR, currentBalanceDCR)
	} else {
		infoMsg = "Processing your speech request (billing disabled)..."
	}
	s.bot.SendPM(ctx, req.UserNick, infoMsg)

	// 3. Create the appropriate FAL request object using the helper function
	falReq, err := createFalSpeechRequest(req)
	if err != nil {
		// Log error server-side, do not PM the user here.
		// Error will be handled by the command handler.
		// s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error creating speech request: %v", err))
		return &SpeechResult{Success: false, Error: err}, err // Return error to command handler
	}

	// 4. Generate speech using the created request
	audioResp, genErr := s.client.GenerateSpeech(ctx, falReq)
	if genErr != nil {
		// Log error server-side, do not PM the user here.
		// Error will be handled by the command handler.
		// s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Speech generation failed: %v", genErr))
		return &SpeechResult{Success: false, Error: genErr}, genErr // Return error to command handler
	}

	// 5. Check if the audio URL is empty
	if audioResp.AudioURL == "" {
		genErr = fmt.Errorf("received empty audio URL from API")
		// Log error server-side, do not PM the user here.
		// Error will be handled by the command handler.
		// s.bot.SendPM(ctx, req.UserNick, genErr.Error())
		return &SpeechResult{Success: false, Error: genErr}, genErr // Return error to command handler
	}

	// 6. Download and send audio
	successfullySent := false
	if err := s.downloadAndSendAudio(ctx, req.UserNick, audioResp.AudioURL, req.ModelName); err != nil {
		// Log download/send error server-side, do not PM the user here.
		fmt.Printf("ERROR [SpeechService] User %s: Failed to download/send audio: %v\n", req.UserNick, err)
		// s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Failed to send audio file: %v", err))
		// Continue but mark as not sent for billing purposes
	} else {
		successfullySent = true
	}

	// 7. Perform Billing *only if* enabled and audio was sent successfully
	var chargedDCR float64
	var finalBalanceDCR float64 = currentBalanceDCR // Use pre-deduction balance (balance from CheckBalance)
	var billingAttempted bool = false
	var billingSucceeded bool = false

	if s.billingEnabled && successfullySent {
		billingAttempted = true
		deductChargedDCR, deductNewBalance, deductErr := utils.DeductBalance(ctx, s.dbManager, req.UserID[:], req.PriceUSD, s.debug, s.billingEnabled)
		if deductErr != nil {
			s.bot.SendPM(ctx, req.UserNick, fmt.Sprintf("Error processing payment after sending audio: %v. Please contact support.", deductErr))
			finalBalanceDCR = currentBalanceDCR // Use pre-deduction balance
		} else {
			billingSucceeded = true
			chargedDCR = deductChargedDCR
			finalBalanceDCR = deductNewBalance
		}
	} else if !s.billingEnabled {
		// fmt.Printf("INFO: Billing disabled. No charge for speech for user %s.\n", req.UserNick) // Already Removed
	} else {
		// Billing enabled, but not sent successfully
		// fmt.Printf("INFO: Audio not sent successfully for user %s. No billing occurred.\n", req.UserNick) // Removed
	}

	// 8. Send final confirmation
	finalMessage := "Finished processing speech request.\n\n"
	if !successfullySent {
		finalMessage = "Speech generation completed, but failed to send the result.\n\n"
	}

	if s.billingEnabled {
		if billingAttempted && billingSucceeded {
			finalMessage += fmt.Sprintf("💰 Billing Information:\n• Charged: %.8f DCR ($%.2f USD)\n• New Balance: %.8f DCR",
				chargedDCR, req.PriceUSD, finalBalanceDCR)
		} else if billingAttempted && !billingSucceeded {
			finalMessage += fmt.Sprintf("⚠️ Billing failed after sending audio. Your balance remains %.8f DCR. Please contact support.", finalBalanceDCR)
		} else {
			finalMessage += fmt.Sprintf("No charge was applied. Your balance remains %.8f DCR.", finalBalanceDCR)
		}
	} else {
		finalMessage += "Billing is disabled. No charge was applied."
	}

	if err := s.bot.SendPM(ctx, req.UserNick, finalMessage); err != nil {
		// fmt.Printf("ERROR: Failed to send final confirmation message (speech) to %s: %v\n", req.UserNick, err) // Removed
	}

	// Return overall success based on generation, even if sending/billing failed
	return &SpeechResult{
		AudioURL: audioResp.AudioURL,
		Success:  true, // Represents successful generation
	}, nil
}

// downloadAndSendAudio fetches audio, saves to temp file, and sends via SendFile
func (s *SpeechService) downloadAndSendAudio(ctx context.Context, userNick string, audioURL string, modelName string) error {
	// Determine filename/extension (use info from response if available, else default)
	// For now, defaulting to mp3 based on minimax default format
	// A more robust approach would pass content_type from fal.AudioResponse
	fileExtension := ".mp3"
	fileNamePrefix := "speech-" + modelName + "-"

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", fileNamePrefix+"*"+fileExtension)
	if err != nil {
		return fmt.Errorf("failed to create temp audio file: %v", err)
	}
	// Ensure the temp file is removed regardless of success/failure
	defer func() {
		err := os.Remove(tmpFile.Name())
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("WARN: Failed to remove temp audio file %s: %v\n", tmpFile.Name(), err)
		}
	}()

	// Fetch the audio data
	audioRespHTTP, err := http.Get(audioURL)
	if err != nil {
		return fmt.Errorf("failed to fetch audio: %v", err)
	}
	defer audioRespHTTP.Body.Close()

	if audioRespHTTP.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch audio: status code %d", audioRespHTTP.StatusCode)
	}

	// Copy the downloaded data to the temp file
	_, err = io.Copy(tmpFile, audioRespHTTP.Body)
	if err != nil {
		// Attempt to close file before returning error
		_ = tmpFile.Close()
		return fmt.Errorf("failed to save audio to temp file: %v", err)
	}

	// Close the file before sending
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp audio file: %v", err)
	}

	// Send the file to the user
	if err := s.bot.SendFile(ctx, userNick, tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to send audio file: %v", err)
	}

	return nil
}

// createFalSpeechRequest constructs the appropriate fal.Model request struct based on the internal SpeechRequest.
func createFalSpeechRequest(req *SpeechRequest) (interface{}, error) {
	var falReq interface{}

	// Create the specific fal request based on the model name
	switch req.ModelName {
	case "minimax-tts/text-to-speech":
		falReq = &fal.MinimaxTTSRequest{
			BaseSpeechRequest: fal.BaseSpeechRequest{
				Text:     req.Text,
				Progress: req.Progress,
			},
			VoiceID: req.VoiceID, // Use parsed voice ID
			// Copy parsed options
			Speed:      req.Speed,
			Vol:        req.Vol,
			Pitch:      req.Pitch,
			Emotion:    req.Emotion,
			SampleRate: req.SampleRate,
			Bitrate:    req.Bitrate,
			Format:     req.Format,
			Channel:    req.Channel,
		}
	// Add cases for other specific speech models here
	default:
		return nil, fmt.Errorf("unsupported or unhandled model for specific FAL speech request creation: %s", req.ModelName)
	}
	return falReq, nil
}
