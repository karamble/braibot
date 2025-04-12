package audio

// DeviceType represents the type of audio device
type DeviceType string

const (
	// DeviceTypeCapture represents a capture (recording) device
	DeviceTypeCapture DeviceType = "capture"
	// DeviceTypePlayback represents a playback device
	DeviceTypePlayback DeviceType = "playback"
)

// Device represents an audio device
type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// Devices represents a collection of audio devices
type Devices struct {
	Playback []Device `json:"playback"`
	Capture  []Device `json:"capture"`
}

// RecordInfo contains information about a recorded audio segment
type RecordInfo struct {
	SampleCount int `json:"sample_count"`
	DurationMs  int `json:"duration_ms"`
	EncodedSize int `json:"encoded_size"`
	PacketCount int `json:"packet_count"`
}
