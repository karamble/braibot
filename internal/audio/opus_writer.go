package audio

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	opusIdSig      = "OpusHead"
	opusCommentSig = "OpusTags"
)

// OpusPacket represents an encoded Opus packet
type OpusPacket []byte

type opusWriter struct {
	ogg *OggWriter

	totalPCMSamples uint64
	pageIndex       uint32
}

// NewOpusWriter creates a new Opus writer
func NewOpusWriter(out io.Writer) (*opusWriter, error) {
	oggWriter := NewOggWriter(out)

	writer := &opusWriter{
		ogg: oggWriter,
	}

	err := writer.writeHeaders()
	if err != nil {
		return nil, err
	}

	return writer, nil
}

// writeHeaders writes the Opus headers to the OGG container
func (w *opusWriter) writeHeaders() error {
	// Write Opus identification header
	idHeader := make([]byte, 19)
	copy(idHeader[0:], opusIdSig)
	idHeader[8] = 1 // Version
	idHeader[9] = 2 // Channels

	binary.LittleEndian.PutUint16(idHeader[10:], 0)     // pre-skip
	binary.LittleEndian.PutUint32(idHeader[12:], 48000) // sample rate
	binary.LittleEndian.PutUint16(idHeader[16:], 0)     // output gain
	idHeader[18] = 0                                    // mono or stereo

	idPage := w.ogg.NewPage(idHeader, 0, w.pageIndex)
	idPage.IsFirstPage = true
	err := w.ogg.WritePage(idPage)
	if err != nil {
		return err
	}
	w.pageIndex++

	// Write Opus comment header
	commentHeader := make([]byte, 25)
	copy(commentHeader[0:], opusCommentSig)
	binary.LittleEndian.PutUint32(commentHeader[8:], 9)  // vendor name length
	copy(commentHeader[12:], "braibot")                  // vendor name
	binary.LittleEndian.PutUint32(commentHeader[21:], 0) // comment list length

	commentPage := w.ogg.NewPage(commentHeader, 0, w.pageIndex)
	err = w.ogg.WritePage(commentPage)
	if err == nil {
		w.pageIndex++
	}
	return err
}

// WritePacket writes an Opus packet to the OGG container
func (w *opusWriter) WritePacket(p []byte, pcmSamples uint64, isLast bool) error {
	if len(p) > 255*255 {
		// Such a large payload requires splitting a single packet into
		// multiple ogg pages.
		return fmt.Errorf("packet splitting not supported")
	}
	granule := w.totalPCMSamples + pcmSamples
	w.totalPCMSamples += pcmSamples
	page := w.ogg.NewPage(p, granule, w.pageIndex)
	page.IsLastPage = isLast
	w.pageIndex++

	return w.ogg.WritePage(page)
}
