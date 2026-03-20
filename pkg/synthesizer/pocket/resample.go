package pocket

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/dharmab/skyeye/pkg/pcm/rate"
	"github.com/martinlindhe/unit"
	"github.com/zaf/resample"
)

// downsample resamples F32LE PCM audio from sourceRate down to 16kHz wideband.
func downsample(samples []float32, sourceRate unit.Frequency) ([]float32, error) {
	// Convert []float32 to []byte (F32LE)
	input := make([]byte, len(samples)*4)
	for i, s := range samples {
		binary.LittleEndian.PutUint32(input[i*4:], math.Float32bits(s))
	}

	const channels = 1
	var buf bytes.Buffer
	resampler, err := resample.New(&buf, sourceRate.Hertz(), rate.Wideband.Hertz(), channels, resample.F32, resample.LowQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create resampler: %w", err)
	}
	defer resampler.Close()

	if _, err = resampler.Write(input); err != nil {
		return nil, fmt.Errorf("failed to resample synthesized audio: %w", err)
	}

	// Convert []byte (F32LE) back to []float32
	output := buf.Bytes()
	result := make([]float32, len(output)/4)
	for i := range result {
		result[i] = math.Float32frombits(binary.LittleEndian.Uint32(output[i*4:]))
	}
	return result, nil
}
