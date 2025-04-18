package speech

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/karamble/braibot/internal/database"
	"github.com/karamble/braibot/internal/utils"
	"github.com/karamble/braibot/pkg/fal"
	kit "github.com/vctt94/bisonbotkit"
)

// SpeechService handles speech generation
type SpeechService struct {
	client    *fal.Client
	dbManager *database.DBManager
	bot       *kit.Bot
	debug     bool
}

// NewSpeechService creates a new SpeechService
func NewSpeechService(client *fal.Client, dbManager *database.DBManager, bot *kit.Bot, debug bool) *SpeechService {
	return &SpeechService{
		client:    client,
		dbManager: dbManager,
		bot:       bot,
		debug:     debug,
	}
}

// GenerateSpeech generates speech based on the internal request
func (s *SpeechService) GenerateSpeech(ctx context.Context, req *SpeechRequest) (*SpeechResult, error) {
	// 1. Check balance (No validation needed here as options are simple for now)
	pm := types.ReceivedPM{
		Nick: req.UserNick,
		Uid:  req.UserID[:],
	}
	billingResult, billingErr := utils.CheckAndProcessBilling(ctx, s.bot, s.dbManager, pm, req.PriceUSD, s.debug)
	if billingErr != nil {
		return &SpeechResult{Success: false, Error: billingErr}, billingErr
	}
	if !billingResult.Success {
		return &SpeechResult{Success: false, Error: fmt.Errorf(billingResult.ErrorMessage)}, nil
	}

	// 2. Send initial message
	s.bot.SendPM(ctx, req.UserNick, "Processing your speech request.")

	// 3. Create the appropriate FAL request object using the helper function
	falReq, err := createFalSpeechRequest(req)
	if err != nil {
		// Optional: Refund logic
		return &SpeechResult{Success: false, Error: err}, err
	}

	// 4. Generate speech using the created request
	audioResp, err := s.client.GenerateSpeech(ctx, falReq)
	if err != nil {
		// Optional: Refund logic
		return &SpeechResult{Success: false, Error: err}, err
	}

	// 5. Check if the audio URL is empty
	if audioResp.AudioURL == "" {
		err = fmt.Errorf("received empty audio URL from API")
		// Optional: Refund logic
		return &SpeechResult{Success: false, Error: err}, err
	}

	// 6. Download and send audio
	if err := s.downloadAndSendAudio(ctx, req.UserNick, audioResp.AudioURL, req.ModelName); err != nil {
		// Optional: Refund logic, although user was likely charged already
		return &SpeechResult{Success: false, Error: err}, err
	}

	// 7. Send billing info
	if err := utils.SendBillingMessage(ctx, s.bot, pm, billingResult); err != nil {
		// Log error, but don't fail the whole operation
		fmt.Printf("ERROR: Failed to send billing message to %s: %v\n", req.UserNick, err)
	}

	return &SpeechResult{
		AudioURL: audioResp.AudioURL,
		Success:  true,
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
