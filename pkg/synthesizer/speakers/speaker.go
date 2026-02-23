// Package speakers contains interfaces and implementations for text-to-speech speakers.
package speakers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/dharmab/skyeye/pkg/pcm/rate"
	"github.com/martinlindhe/unit"
	"github.com/zaf/resample"
)

// Speaker provides text-to-speech.
type Speaker interface {
	// Say returns F32LE PCM audio for the given text.
	Say(context.Context, string) ([]float32, error)
}

func downsample(sample []byte, sourceRate unit.Frequency) ([]byte, error) {
	const channels = 1
	var buf bytes.Buffer
	resampler, err := resample.New(&buf, sourceRate.Hertz(), rate.Wideband.Hertz(), channels, resample.I16, resample.LowQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create resampler: %w", err)
	}
	defer resampler.Close()

	_, err = resampler.Write(sample)
	if err != nil {
		return nil, fmt.Errorf("failed to resample synthesized audio: %w", err)
	}
	return buf.Bytes(), nil
}
