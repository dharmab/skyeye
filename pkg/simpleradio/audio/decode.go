package audio

import (
	"fmt"
	"time"

	"github.com/pion/opus"
)

const (
	frameLength = 40 * time.Millisecond
	sampleRate  = 16000
	channels    = 1
)

var frameSize = channels * frameLength.Milliseconds() * sampleRate / 1000

// decode the given Opus audio from SRS to F32LE PCM
func decode(decoder opus.Decoder, b []byte) (Audio, error) {
	pcm := make([]float32, frameSize)
	_, _, err := decoder.DecodeFloat32(b, pcm)
	if err != nil {
		return nil, fmt.Errorf("failed to decode audio: %w", err)
	}
	return pcm, nil
}
