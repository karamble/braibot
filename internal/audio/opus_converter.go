package audio

import (
	"bytes"
	"fmt"

	"github.com/companyzero/gopus"
)

// ConvertPCMToOpus converts PCM audio data to Opus format with proper OGG container
func ConvertPCMToOpus(pcmData []byte) ([]byte, error) {
	// Create Opus encoder
	const sampleRate = 24000 // Opus supported sample rate
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
	opusWriter, err := NewOpusWriter(&oggBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus writer: %v", err)
	}

	// Opus frame size must be one of: 120, 240, 480, 960, 1920, 2880 samples
	// Using 960 samples (40ms at 24kHz) for good quality and reasonable latency
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
