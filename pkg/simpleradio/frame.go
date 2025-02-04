package simpleradio

import (
	"fmt"
	"time"

	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/martinlindhe/unit"
	"gopkg.in/hraban/opus.v2"
)

const (
	// frameLength is the length of an Opus frame sent by SRS.
	frameLength = 40 * time.Millisecond
	// sampleRate is the sample rate of the audio data sent by SRS.
	sampleRate = 16 * unit.Kilohertz // Wideband
	// channels is the number of channels in the audio data sent by SRS.
	channels = 1 // Mono
	// encodingBufferSize is the size of the buffer used to encode audio data. The buffer size may effect bitrate.
	encodingBufferSize = 1024
)

// frameSize is the Opus frame size used in SRS voice packets.
var frameSize = channels * frameLength.Milliseconds() * int64(sampleRate.Kilohertz())

// decodeFrame decodes the given Opus frame(s) into F32LE PCM audio data.
func decodeFrame(decoder *opus.Decoder, b []byte) ([]float32, error) {
	f32le := make([]float32, frameSize)
	n, err := decoder.DecodeFloat32(b, f32le)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Opus audio: %w", err)
	}
	f32le = f32le[:n*channels]
	return f32le, nil
}

// encodeFrame encodes the given F32LE PCM audio data into an Opus frame.
func encodeFrame(encoder *opus.Encoder, f32le []float32) ([]byte, error) {
	b := make([]byte, encodingBufferSize)
	n, err := encoder.Encode(pcm.F32toS16LE(f32le), b)
	if err != nil {
		return b, fmt.Errorf("failed to encode Opus audio: %w", err)
	}
	b = b[:n]
	return b, nil
}
