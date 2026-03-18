// Package voice provides the default reference audio for Pocket TTS voice cloning.
// This package has no CGO dependencies and can be built with CGO_ENABLED=0.
package voice

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	_ "embed"
)

// DefaultVoice is the embedded default reference WAV file for voice cloning.
//
//go:embed default.wav
var DefaultVoice []byte

// DecodeWAV decodes a 16-bit PCM mono WAV file into float32 samples and its sample rate.
// The input must be a valid WAV file with 16-bit signed PCM encoding and exactly 1 channel.
func DecodeWAV(data []byte) (samples []float32, sampleRate int, err error) {
	if len(data) < 44 {
		return nil, 0, errors.New("WAV data too short for header")
	}

	// Verify RIFF header
	if string(data[0:4]) != "RIFF" {
		return nil, 0, errors.New("missing RIFF header")
	}
	if string(data[8:12]) != "WAVE" {
		return nil, 0, errors.New("missing WAVE format identifier")
	}

	// Find fmt chunk
	offset := 12
	var fmtFound bool
	var audioFormat uint16
	var numChannels uint16
	var bitsPerSample uint16

	for offset+8 <= len(data) {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))

		if chunkID == "fmt " {
			if chunkSize < 16 || offset+24 > len(data) {
				return nil, 0, errors.New("fmt chunk too small")
			}
			audioFormat = binary.LittleEndian.Uint16(data[offset+8 : offset+10])
			numChannels = binary.LittleEndian.Uint16(data[offset+10 : offset+12])
			sampleRate = int(binary.LittleEndian.Uint32(data[offset+12 : offset+16]))
			bitsPerSample = binary.LittleEndian.Uint16(data[offset+22 : offset+24])
			fmtFound = true
		}

		if chunkID == "data" {
			if !fmtFound {
				return nil, 0, errors.New("data chunk before fmt chunk")
			}
			if audioFormat != 1 {
				return nil, 0, fmt.Errorf("unsupported audio format %d (expected 1 = PCM)", audioFormat)
			}
			if numChannels != 1 {
				return nil, 0, fmt.Errorf("unsupported channel count %d (expected 1 = mono)", numChannels)
			}
			if bitsPerSample != 16 {
				return nil, 0, fmt.Errorf("unsupported bits per sample %d (expected 16)", bitsPerSample)
			}

			dataStart := offset + 8
			dataEnd := min(dataStart+chunkSize, len(data))
			pcmData := data[dataStart:dataEnd]

			numSamples := len(pcmData) / 2
			samples = make([]float32, numSamples)
			for i := range numSamples {
				s := int16(binary.LittleEndian.Uint16(pcmData[i*2 : i*2+2]))
				samples[i] = float32(s) / math.MaxInt16
			}
			return samples, sampleRate, nil
		}

		// Advance to next chunk (chunks are word-aligned)
		offset += 8 + chunkSize
		if chunkSize%2 != 0 {
			offset++
		}
	}

	if !fmtFound {
		return nil, 0, errors.New("fmt chunk not found")
	}
	return nil, 0, errors.New("data chunk not found")
}
