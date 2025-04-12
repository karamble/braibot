package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/companyzero/bisonrelay/clientrpc/types"
	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/companyzero/gopus"
	"github.com/decred/dcrd/dcrutil/v4"
	"github.com/karamble/braibot/internal/audio"
	kit "github.com/vctt94/bisonbotkit"
	"github.com/vctt94/bisonbotkit/config"
	"github.com/vctt94/bisonbotkit/logging"
	"github.com/vctt94/bisonbotkit/utils"
)

var (
	flagAppRoot = flag.String("approot", "~/.braibot", "Path to application data directory")
	// currentModel = "fast-sdxl" // Default model
	debug     = true     // Set to true for debugging
	dbManager *DBManager // Database manager for user balances
)

// Define a struct for the model details
type Model struct {
	Name        string  // Name of the model
	Description string  // Description of the model
	Price       float64 // Price per picture in USD
}

// Separate model maps for each command type
var text2imageModels = map[string]Model{
	"fast-sdxl": {
		Name:        "fast-sdxl",
		Description: "Fast model for generating images quickly.",
		Price:       0.02,
	},
	"hidream-i1-full": {
		Name:        "hidream-i1-full",
		Description: "High-quality model for detailed images.",
		Price:       0.10,
	},
	"hidream-i1-dev": {
		Name:        "hidream-i1-dev",
		Description: "Development version of the HiDream model.",
		Price:       0.06,
	},
	"hidream-i1-fast": {
		Name:        "hidream-i1-fast",
		Description: "Faster version of the HiDream model.",
		Price:       0.03,
	},
	"flux-pro/v1.1": {
		Name:        "flux-pro/v1.1",
		Description: "Professional model for high-end image generation.",
		Price:       0.08,
	},
	"flux-pro/v1.1-ultra": {
		Name:        "flux-pro/v1.1-ultra",
		Description: "Ultra version of the professional model.",
		Price:       0.12,
	},
	"flux/schnell": {
		Name:        "flux/schnell",
		Description: "Quick model for rapid image generation.",
		Price:       0.02,
	},
}

var text2speechModels = map[string]Model{
	"minimax-tts/text-to-speech": {
		Name:        "minimax-tts/text-to-speech",
		Description: "Text-to-speech model for converting text to audio. $0.10 per 1000 characters.",
		Price:       0.10,
	},
}

// Map to hold the current model for each command
var currentModels = map[string]string{
	"text2image":  "fast-sdxl",                  // Default model for text2image
	"text2speech": "minimax-tts/text-to-speech", // Default model for text2speech
}

// Command represents a bot command
type Command struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error
}

// FalResponse represents the response from Fal.ai API
type FalResponse struct {
	Status        string `json:"status,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	ResponseURL   string `json:"response_url,omitempty"`
	StatusURL     string `json:"status_url,omitempty"`
	CancelURL     string `json:"cancel_url,omitempty"`
	QueuePosition int    `json:"queue_position,omitempty"`
	Logs          []struct {
		Message   string `json:"message"`
		Level     string `json:"level"`
		Source    string `json:"source"`
		Timestamp string `json:"timestamp"`
	} `json:"logs,omitempty"`
	Response struct {
		Images []struct {
			URL         string `json:"url"`
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			ContentType string `json:"content_type"`
		} `json:"images"`
	} `json:"response,omitempty"`
}

// Available commands
var commands map[string]Command

const oggSig = "OggS"

type OggHeader struct {
	Version     uint8
	IsContinued bool
	IsFirstPage bool
	IsLastPage  bool

	GranulePosition uint64
	BitstreamSerial uint32
	PageSequence    uint32
	CrcChecksum     uint32

	PageSegments uint8
	SegmentTable []uint8
}

type OggPage struct {
	OggHeader
	Segments [][]byte

	// Size of all segments in bytes
	SegmentTotal int
}

var checksumTable = crcChecksum()

type oggWriter struct {
	w      io.Writer
	serial uint32
}

func newOggWriter(out io.Writer) *oggWriter {
	return &oggWriter{
		w:      out,
		serial: rand.Uint32(),
	}
}

func (o *oggWriter) WritePage(p OggPage) error {
	headerSize := 27 + int(p.PageSegments)
	totalSize := headerSize + p.SegmentTotal

	buf := make([]byte, totalSize)
	headerType := uint8(0x0)
	if p.IsContinued {
		headerType = headerType | 0x1
	}
	if p.IsFirstPage {
		headerType = headerType | 0x2
	}
	if p.IsLastPage {
		headerType = headerType | 0x4
	}

	copy(buf[0:], oggSig)
	buf[4] = p.Version
	buf[5] = headerType

	binary.LittleEndian.PutUint64(buf[6:], p.GranulePosition)
	binary.LittleEndian.PutUint32(buf[14:], p.BitstreamSerial)
	binary.LittleEndian.PutUint32(buf[18:], p.PageSequence)
	// compute checksum later

	buf[26] = p.PageSegments
	for i, s := range p.SegmentTable {
		buf[27+i] = s
	}

	idx := headerSize
	for i, s := range p.Segments {
		copy(buf[idx:], s)
		idx += int(p.SegmentTable[i])
	}

	var checksum uint32
	for i := range buf {
		checksum = (checksum << 8) ^ checksumTable[byte(checksum>>24)^buf[i]]
	}
	binary.LittleEndian.PutUint32(buf[22:], checksum)

	_, err := o.w.Write(buf)
	return err
}

// partions a slice of bytes into units no bigger than 255
func partition(p []byte) ([]uint8, [][]byte) {
	segCountHint := len(p)/255 + 1
	st := make([]uint8, 0, segCountHint)
	s := make([][]byte, 0, segCountHint)

	for len(p) > 255 {
		st = append(st, 255)
		s = append(s, p[:255])
		p = p[255:]
	}

	st = append(st, uint8(len(p)))
	s = append(s, p)

	// packet of exactly 255 bytes is terminated by lacing value of 0
	if len(p) == 255 {
		st = append(st, 0)
		s = append(s, []byte{})
	}
	return st, s
}

func (o *oggWriter) NewPage(payload []byte, granulePosition uint64, pageSeqence uint32) OggPage {
	segTable, segments := partition(payload)
	total := len(payload)

	return OggPage{
		OggHeader: OggHeader{
			Version:         0,
			GranulePosition: granulePosition,
			BitstreamSerial: o.serial,
			PageSequence:    pageSeqence,

			PageSegments: uint8(len(segTable)),
			SegmentTable: segTable,
		},
		Segments:     segments,
		SegmentTotal: total,
	}
}

func (o *oggWriter) Finish(granulePosition uint64, pageSeqence uint32) error {
	page := o.NewPage([]byte{}, granulePosition, pageSeqence)
	page.IsLastPage = true
	return o.WritePage(page)
}

// https://github.com/pion/webrtc/blob/67826b19141ec9e6f1002a2267008a016a118934/pkg/media/oggwriter/oggwriter.go#L245-L261
func crcChecksum() *[256]uint32 {
	var table [256]uint32
	const poly = 0x04c11db7

	for i := range table {
		r := uint32(i) << 24
		for j := 0; j < 8; j++ {
			if (r & 0x80000000) != 0 {
				r = (r << 1) ^ poly
			} else {
				r <<= 1
			}
			table[i] = (r & 0xffffffff)
		}
	}
	return &table
}

func init() {
	commands = map[string]Command{
		"help": {
			Name:        "help",
			Description: "Shows this help message",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				helpMsg := "| Command | Description |\n| -------- | ----------- |\n"
				for _, cmd := range commands {
					helpMsg += fmt.Sprintf("| !%s | %s |\n", cmd.Name, cmd.Description)
				}
				return bot.SendPM(ctx, pm.Nick, helpMsg)
			},
		},
		"listmodels": {
			Name:        "listmodels",
			Description: "Lists available models for a specific command. Usage: !listmodels [command]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) == 0 {
					return bot.SendPM(ctx, pm.Nick, "Please specify a command. Usage: !listmodels [command]")
				}

				commandName := strings.ToLower(args[0])

				var modelList string
				var models map[string]Model

				switch commandName {
				case "text2image":
					models = text2imageModels
					modelList = "Available models for text2image:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
				case "text2speech":
					models = text2speechModels
					modelList = "Available models for text2speech:\n| Model | Description | Price |\n| ----- | ----------- | ----- |\n"
				default:
					return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image' or 'text2speech'.")
				}

				for _, model := range models {
					modelList += fmt.Sprintf("| %s | %s | $%.2f |\n", model.Name, model.Description, model.Price)
				}

				return bot.SendPM(ctx, pm.Nick, modelList)
			},
		},
		"setmodel": {
			Name:        "setmodel",
			Description: "Sets the model to use for specified commands. Usage: !setmodel [command] [modelname]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) < 2 {
					return bot.SendPM(ctx, pm.Nick, "Please specify a command and a model name. Usage: !setmodel [command] [modelname]")
				}
				commandName := args[0]
				modelName := args[1]

				// Check if the command is valid
				if _, exists := commands[commandName]; !exists {
					return bot.SendPM(ctx, pm.Nick, "Invalid command name. Use !listmodels to see available commands.")
				}

				// Check if the model is valid for the specific command
				var models map[string]Model
				switch commandName {
				case "text2image":
					models = text2imageModels
				case "text2speech":
					models = text2speechModels
				default:
					return bot.SendPM(ctx, pm.Nick, "Invalid command. Use 'text2image' or 'text2speech'.")
				}

				if _, exists := models[modelName]; exists {
					currentModels[commandName] = modelName
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Model for %s set to: %s", commandName, modelName))
				}

				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Invalid model name for %s. Use !listmodels %s to see available models.", commandName, commandName))
			},
		},
		"text2image": {
			Name:        "text2image",
			Description: "Generates an image from text prompt. Usage: !text2image [prompt]",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) == 0 {
					return bot.SendPM(ctx, pm.Nick, "Please provide a prompt. Usage: !text2image [prompt]")
				}

				prompt := strings.Join(args, " ")

				// Prepare the request
				requestBody, err := json.Marshal(map[string]interface{}{
					"prompt": prompt,
				})
				if err != nil {
					return err
				}

				// Use the current model for text2image
				modelToUse := currentModels["text2image"]

				// Create HTTP request for initial call
				req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://queue.fal.run/fal-ai/%s", modelToUse), bytes.NewBuffer(requestBody))
				if err != nil {
					return err
				}

				// Set headers
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

				// Send request
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				// Read response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				// Parse initial response
				var initialResp FalResponse
				if err := json.Unmarshal(body, &initialResp); err != nil {
					return err
				}

				// Poll until completion
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						// Check status with logs enabled
						statusReq, err := http.NewRequestWithContext(ctx, "GET", initialResp.StatusURL+"?logs=1", nil)
						if err != nil {
							return err
						}
						statusReq.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

						statusResp, err := client.Do(statusReq)
						if err != nil {
							return err
						}

						statusBody, err := io.ReadAll(statusResp.Body)
						statusResp.Body.Close()
						if err != nil {
							return err
						}

						var statusResponse FalResponse
						if err := json.Unmarshal(statusBody, &statusResponse); err != nil {
							return err
						}

						switch statusResponse.Status {
						case "IN_QUEUE":
							// Send queue position update
							bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Your request is in queue. Position: %d", statusResponse.QueuePosition))
							continue
						case "IN_PROGRESS":
							// Log progress if available
							if len(statusResponse.Logs) > 0 {
								bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Processing: %s", statusResponse.Logs[len(statusResponse.Logs)-1].Message))
							}
							continue
						case "COMPLETED":
							// Fetch final response
							finalReq, err := http.NewRequestWithContext(ctx, "GET", initialResp.ResponseURL, nil)
							if err != nil {
								return err
							}
							finalReq.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

							finalResp, err := client.Do(finalReq)
							if err != nil {
								return err
							}
							defer finalResp.Body.Close()

							// Check the status code
							if finalResp.StatusCode != http.StatusOK {
								body, _ := io.ReadAll(finalResp.Body) // Read the body for logging
								return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching final response: %s. Body: %s", finalResp.Status, string(body)))
							}

							finalBody, err := io.ReadAll(finalResp.Body)
							if err != nil {
								return err
							}

							// Debug output
							if debug {
								fmt.Printf("Final Response Body: %s\n", string(finalBody))
							}

							// Unmarshal the final response
							var finalResponse struct {
								Images []struct {
									URL         string `json:"url"`
									Width       int    `json:"width"`
									Height      int    `json:"height"`
									ContentType string `json:"content_type"`
								} `json:"images"`
								Timings struct {
									Inference float64 `json:"inference"`
								} `json:"timings"`
								Seed            json.Number `json:"seed"`
								HasNSFWConcepts []bool      `json:"has_nsfw_concepts"`
								Prompt          string      `json:"prompt"`
							}
							if err := json.Unmarshal(finalBody, &finalResponse); err != nil {
								return err
							}

							// Assuming the first image is the one we want to send
							if len(finalResponse.Images) > 0 {
								imageURL := finalResponse.Images[0].URL
								// Fetch the image data
								imgResp, err := http.Get(imageURL)
								if err != nil {
									return err
								}
								defer imgResp.Body.Close()

								imgData, err := io.ReadAll(imgResp.Body)
								if err != nil {
									return err
								}

								// Encode the image data to base64
								encodedImage := base64.StdEncoding.EncodeToString(imgData)

								// Determine the image type from ContentType
								var imageType string
								switch finalResponse.Images[0].ContentType {
								case "image/jpeg":
									imageType = "image/jpeg"
								case "image/png":
									imageType = "image/png"
								case "image/webp":
									imageType = "image/webp"
								default:
									imageType = "image/jpeg" // Fallback to jpeg if unknown
								}

								// Create the message with embedded image, using the user's prompt as the alt text
								message := fmt.Sprintf("--embed[alt=%s,type=%s,data=%s]--", url.QueryEscape(prompt), imageType, encodedImage)
								return bot.SendPM(ctx, pm.Nick, message)
							} else {
								return bot.SendPM(ctx, pm.Nick, "No images were generated.")
							}
						case "FAILED":
							// Send the complete raw response body as PM
							responseMessage := fmt.Sprintf("Failed to generate image. Complete response: %s", string(statusBody))
							return bot.SendPM(ctx, pm.Nick, responseMessage)
						default:
							// Still processing, continue polling
							continue
						}
					}
				}
			},
		},
		"balance": {
			Name:        "balance",
			Description: "Shows your current balance",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				// Convert UID to string ID for database, just like in tip handler
				var userID zkidentity.ShortID
				userID.FromBytes(pm.Uid)
				userIDStr := userID.String()

				// Get balance from database using the proper ID
				balance, err := dbManager.GetBalance(userIDStr)
				if err != nil {
					return fmt.Errorf("failed to get balance: %v", err)
				}

				// Convert to DCR using dcrutil, same as in tip handler
				dcrBalance := dcrutil.Amount(balance / 1e3).ToCoin()

				// Send balance message
				return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Your current balance is %.8f DCR", dcrBalance))
			},
		},
		"rate": {
			Name:        "rate",
			Description: "Shows current DCR exchange rate in USD and BTC",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				// Send a status message to indicate we're fetching rates
				bot.SendPM(ctx, pm.Nick, "Fetching current exchange rates...")

				// Create HTTP client with timeout
				client := &http.Client{
					Timeout: 10 * time.Second,
				}

				// Make request to dcrdata API
				resp, err := client.Get("https://explorer.dcrdata.org/api/exchangerate")
				if err != nil {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching rates: %v", err))
				}
				defer resp.Body.Close()

				// Check status code
				if resp.StatusCode != http.StatusOK {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error: API returned status %d", resp.StatusCode))
				}

				var rates struct {
					DCRPrice float64 `json:"dcrPrice"`
					BTCPrice float64 `json:"btcPrice"`
					Time     int64   `json:"time"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
					return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error parsing rates: %v", err))
				}

				// Convert timestamp to human readable format
				timeStr := time.Unix(rates.Time, 0).Format(time.RFC1123)

				// Format the response with more detailed information
				message := fmt.Sprintf("Current DCR Exchange Rates (as of %s):\n"+
					"USD: $%.2f\n"+
					"BTC: %.8f BTC\n"+
					"Source: dcrdata",
					timeStr, rates.DCRPrice, rates.BTCPrice)

				return bot.SendPM(ctx, pm.Nick, message)
			},
		},
		"text2speech": {
			Name:        "text2speech",
			Description: "Converts text to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl",
			Handler: func(ctx context.Context, bot *kit.Bot, cfg *config.BotConfig, pm types.ReceivedPM, args []string) error {
				if len(args) == 0 {
					return bot.SendPM(ctx, pm.Nick, "Please provide text to convert to speech. Usage: !text2speech [voice_id] [text] - voice_id is optional, defaults to Wise_Woman. Available voices: Wise_Woman, Friendly_Person, Inspirational_girl, Deep_Voice_Man, Calm_Woman, Casual_Guy, Lively_Girl, Patient_Man, Young_Knight, Determined_Man, Lovely_Girl, Decent_Boy, Imposing_Manner, Elegant_Man, Abbess, Sweet_Girl_2, Exuberant_Girl")
				}

				var text string
				var voiceID string = "Wise_Woman" // Default voice

				// Check if first argument might be a voice ID
				if len(args) > 1 {
					possibleVoiceID := args[0]
					// List of valid voice IDs
					validVoices := []string{
						"Wise_Woman", "Friendly_Person", "Inspirational_girl",
						"Deep_Voice_Man", "Calm_Woman", "Casual_Guy",
						"Lively_Girl", "Patient_Man", "Young_Knight",
						"Determined_Man", "Lovely_Girl", "Decent_Boy",
						"Imposing_Manner", "Elegant_Man", "Abbess",
						"Sweet_Girl_2", "Exuberant_Girl",
					}

					// Check if the first argument is a valid voice ID
					isVoiceID := false
					for _, v := range validVoices {
						if v == possibleVoiceID {
							isVoiceID = true
							voiceID = possibleVoiceID
							text = strings.Join(args[1:], " ")
							break
						}
					}

					if !isVoiceID {
						// If not a voice ID, use all args as text
						text = strings.Join(args, " ")
					}
				} else {
					text = strings.Join(args, " ")
				}

				// Prepare the request
				requestBody, err := json.Marshal(map[string]interface{}{
					"text": text,
					"voice_setting": map[string]interface{}{
						"voice_id": voiceID,
					},
					"audio_setting": map[string]interface{}{
						"format":      "pcm",
						"sample_rate": 44100, // Highest supported sample rate (integer)
						"channel":     1,     // Mono audio (integer)
					},
				})
				if err != nil {
					return err
				}

				// Create HTTP request for initial call
				req, err := http.NewRequestWithContext(ctx, "POST", "https://queue.fal.run/fal-ai/minimax-tts/text-to-speech", bytes.NewBuffer(requestBody))
				if err != nil {
					return err
				}

				// Set headers
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

				// Send request
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				// Read response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				// Debug output for initial response
				if debug {
					fmt.Printf("Initial Response: %s\n", string(body))
				}

				// Parse initial response to get request ID
				var initialResp struct {
					RequestID string `json:"request_id"`
				}
				if err := json.Unmarshal(body, &initialResp); err != nil {
					return err
				}

				if initialResp.RequestID == "" {
					return bot.SendPM(ctx, pm.Nick, "Error: No request ID received from the API")
				}

				// Poll until completion
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-ticker.C:
						// Check status using the request ID
						statusURL := fmt.Sprintf("https://queue.fal.run/fal-ai/minimax-tts/requests/%s/status", initialResp.RequestID)
						statusReq, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
						if err != nil {
							return err
						}
						statusReq.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

						statusResp, err := client.Do(statusReq)
						if err != nil {
							return err
						}

						statusBody, err := io.ReadAll(statusResp.Body)
						statusResp.Body.Close()
						if err != nil {
							return err
						}

						// Debug output for status response
						if debug {
							fmt.Printf("Status Response: %s\n", string(statusBody))
						}

						var statusResponse struct {
							Status        string `json:"status"`
							QueuePosition int    `json:"queue_position,omitempty"`
						}
						if err := json.Unmarshal(statusBody, &statusResponse); err != nil {
							return err
						}

						switch statusResponse.Status {
						case "IN_QUEUE":
							// Send queue position update
							bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Your request is in queue. Position: %d", statusResponse.QueuePosition))
							continue
						case "IN_PROGRESS":
							// Send processing status
							bot.SendPM(ctx, pm.Nick, "Your audio is being generated, please be patient")
							continue
						case "COMPLETED":
							// Fetch final response using the request ID
							resultURL := fmt.Sprintf("https://queue.fal.run/fal-ai/minimax-tts/requests/%s", initialResp.RequestID)
							finalReq, err := http.NewRequestWithContext(ctx, "GET", resultURL, nil)
							if err != nil {
								return err
							}
							finalReq.Header.Set("Authorization", "Key "+cfg.ExtraConfig["falapikey"])

							finalResp, err := client.Do(finalReq)
							if err != nil {
								return err
							}
							defer finalResp.Body.Close()

							// Check the status code
							if finalResp.StatusCode != http.StatusOK {
								body, _ := io.ReadAll(finalResp.Body) // Read the body for logging
								return bot.SendPM(ctx, pm.Nick, fmt.Sprintf("Error fetching final response: %s. Body: %s", finalResp.Status, string(body)))
							}

							finalBody, err := io.ReadAll(finalResp.Body)
							if err != nil {
								return err
							}

							// Debug output
							if debug {
								fmt.Printf("Final Response Body: %s\n", string(finalBody))
							}

							// Unmarshal the final response
							var finalResponse struct {
								Audio struct {
									URL         string `json:"url"`
									ContentType string `json:"content_type"`
									FileName    string `json:"file_name"`
									FileSize    int    `json:"file_size"`
								} `json:"audio"`
								DurationMs int `json:"duration_ms"`
							}
							if err := json.Unmarshal(finalBody, &finalResponse); err != nil {
								return err
							}

							// Fetch the audio data
							audioResp, err := http.Get(finalResponse.Audio.URL)
							if err != nil {
								return err
							}
							defer audioResp.Body.Close()

							audioData, err := io.ReadAll(audioResp.Body)
							if err != nil {
								return err
							}

							// Convert PCM to Opus
							opusData, err := convertPCMToOpus(audioData)
							if err != nil {
								return fmt.Errorf("failed to convert audio to Opus: %v", err)
							}

							if debug {
								fmt.Printf("Opus data size before encoding: %d bytes\n", len(opusData))
							}

							// Encode as base64
							encodedAudio := base64.StdEncoding.EncodeToString(opusData)

							if debug {
								fmt.Printf("Base64 encoded size: %d bytes\n", len(encodedAudio))
							}

							// Create the message with embedded audio in BisonRelay format
							message := fmt.Sprintf("--embed[alt=%s,type=%s,filename=%s,data=%s]--",
								url.QueryEscape("Text to Speech"),
								"audio/ogg",
								time.Now().Format("2006-01-02-15_04_05")+"-tts.opus",
								encodedAudio)

							// Debug output
							if debug {
								fmt.Printf("Message for the User of the Audio file: %s\n", message)
							}
							return bot.SendPM(ctx, pm.Nick, message)
						case "FAILED":
							// Send the complete raw response body as PM
							responseMessage := fmt.Sprintf("Failed to generate speech. Complete response: %s", string(statusBody))
							return bot.SendPM(ctx, pm.Nick, responseMessage)
						default:
							// Still processing, continue polling
							continue
						}
					}
				}
			},
		},
	}
}

// isCommand checks if a message is a command (starts with !)
func isCommand(msg string) (string, []string, bool) {
	if !strings.HasPrefix(msg, "!") {
		return "", nil, false
	}

	parts := strings.Fields(msg[1:]) // Remove ! and split
	if len(parts) == 0 {
		return "", nil, false
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	return cmd, args, true
}

// convertPCMToOpus converts PCM audio data to Opus format with proper OGG container
func convertPCMToOpus(pcmData []byte) ([]byte, error) {
	// Create Opus encoder
	const sampleRate = 48000 // Opus standard sample rate
	const channels = 1       // Mono audio
	const bitrate = 64000    // 64kbps bitrate

	enc, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus encoder: %v", err)
	}

	// Set the bitrate
	enc.SetBitrate(bitrate)

	// Create a buffer to store the OGG container
	var oggBuffer bytes.Buffer
	opusWriter, err := audio.NewOpusWriter(&oggBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus writer: %v", err)
	}

	// Opus frame size must be one of: 120, 240, 480, 960, 1920, 2880 samples
	// Using 960 samples (20ms at 48kHz) for good quality and reasonable latency
	const frameSize = 960
	pcmBuffer := make([]int16, frameSize)
	var granulePosition uint64

	// Process PCM data in frames
	for i := 0; i < len(pcmData); i += frameSize * 2 { // *2 because each sample is 2 bytes
		// Calculate how many samples we can process in this frame
		remainingBytes := len(pcmData) - i
		samplesToProcess := frameSize
		if remainingBytes < frameSize*2 {
			samplesToProcess = remainingBytes / 2
		}

		// Clear the buffer for this frame
		for j := range pcmBuffer {
			pcmBuffer[j] = 0
		}

		// Convert bytes to int16 samples
		for j := 0; j < samplesToProcess; j++ {
			if i+j*2+1 < len(pcmData) {
				// Convert little-endian bytes to int16
				pcmBuffer[j] = int16(pcmData[i+j*2]) | int16(pcmData[i+j*2+1])<<8
			}
		}

		// Encode to Opus
		opusFrame := make([]byte, 1275) // Max size for 20ms frame
		encodedData, err := enc.Encode(pcmBuffer, frameSize, opusFrame)
		if err != nil {
			return nil, fmt.Errorf("failed to encode to Opus: %v", err)
		}

		if len(encodedData) > 0 {
			// Update granule position (samples processed)
			granulePosition += uint64(samplesToProcess)

			// Write the Opus frame to the OGG container
			err := opusWriter.WritePacket(encodedData, uint64(samplesToProcess), false)
			if err != nil {
				return nil, fmt.Errorf("failed to write Opus frame: %v", err)
			}
		}
	}

	// Write the final packet
	err = opusWriter.WritePacket([]byte{}, 0, true)
	if err != nil {
		return nil, fmt.Errorf("failed to write final Opus packet: %v", err)
	}

	return oggBuffer.Bytes(), nil
}

func realMain() error {
	flag.Parse()

	// Expand and clean the app root path
	appRoot := utils.CleanAndExpandPath(*flagAppRoot)

	// Initialize database manager
	var err error
	dbManager, err = NewDBManager(appRoot)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	defer dbManager.Close()

	// Initialize logging
	logBackend, err := logging.NewLogBackend(logging.LogConfig{
		LogFile:        filepath.Join(appRoot, "logs", "braibot.log"),
		DebugLevel:     "info",
		MaxLogFiles:    5,
		MaxBufferLines: 1000,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize logging: %v", err)
	}
	defer logBackend.Close()

	// Get a logger for the application
	log := logBackend.Logger("BraiBot")

	// Load bot configuration
	cfg, err := config.LoadBotConfig(appRoot, "braibot.conf")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Create a bidirectional channel for PMs and tips
	pmChan := make(chan types.ReceivedPM)
	tipChan := make(chan types.ReceivedTip)
	tipProgressChan := make(chan types.TipProgressEvent)

	// Set up PM channels/log
	cfg.PMChan = pmChan
	cfg.PMLog = logBackend.Logger("PM")

	// Set up tip channels/logs
	cfg.TipLog = logBackend.Logger("TIP")
	cfg.TipProgressChan = tipProgressChan
	cfg.TipReceivedLog = logBackend.Logger("TIP_RECEIVED")
	cfg.TipReceivedChan = tipChan

	// Create new bot instance
	bot, err := kit.NewBot(cfg, logBackend)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	// Add a goroutine to handle PMs using our bidirectional channel
	go func() {
		for pm := range pmChan {
			log.Infof("Received PM from %s: %s", pm.Nick, pm.Msg.Message)

			// Check if the message is a command
			if cmd, args, isCmd := isCommand(pm.Msg.Message); isCmd {
				if command, exists := commands[cmd]; exists {
					if err := command.Handler(context.Background(), bot, cfg, pm, args); err != nil {
						log.Warnf("Error executing command %s: %v", cmd, err)
					}
				} else {
					// Send error message for unknown command
					bot.SendPM(context.Background(), pm.Nick, "Unknown command. Use !help to see available commands.")
				}
			}
		}
	}()

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add input handling goroutine
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			tokens := strings.SplitN(line, " ", 2)
			if len(tokens) != 2 {
				log.Warn("Invalid format. Use: <nick> <message>")
				continue
			}

			nick, msg := tokens[0], tokens[1]
			if err := bot.SendPM(ctx, nick, msg); err != nil {
				log.Warnf("Failed to send PM: %v", err)
				continue
			}
			log.Infof("-> %s: %s", nick, msg)
		}
		if err := scanner.Err(); err != nil {
			log.Errorf("Error reading input: %v", err)
		}
	}()

	// Handle received tips
	go func() {
		for tip := range tipChan {
			// Convert UID to string ID for database
			var userID zkidentity.ShortID
			userID.FromBytes(tip.Uid)
			userIDStr := userID.String()

			// Update user's balance in the database
			err = dbManager.UpdateBalance(userIDStr, tip.AmountMatoms)
			if err != nil {
				log.Errorf("Failed to update balance: %v", err)
				continue
			}

			// Convert to DCR for display
			dcrAmount := dcrutil.Amount(tip.AmountMatoms / 1e3).ToCoin()

			log.Infof("Tip received: %.8f DCR from %s",
				dcrAmount,
				userIDStr)

			// Send thank you message
			bot.SendPM(ctx, userIDStr,
				fmt.Sprintf("Thank you for the tip of %.8f DCR!", dcrAmount))

			// Acknowledge the tip
			bot.AckTipReceived(ctx, tip.SequenceId)
		}
	}()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Infof("Received shutdown signal: %v", sig)
		bot.Close()
		cancel()
	}()

	// Run the bot with the cancellable context
	if err := bot.Run(ctx); err != nil {
		return fmt.Errorf("bot error: %v", err)
	}

	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
