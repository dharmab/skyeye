// package audiotools contains utilities for converting between different represenations of PCM audio data.

package pcm

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestF32ToS16(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		arg      float32
		expected int16
	}{
		{-1.0, -32767},
		{-0.5, -16383},
		{-0.25, -8191},
		{-0.125, -4095},
		{-0.0625, -2047},
		{-0.03125, -1023},
		{-0.015625, -511},
		{-0.0078125, -255},
		{0.0, 0},
		{0.0078125, 255},
		{0.015625, 511},
		{0.03125, 1023},
		{0.0625, 2047},
		{0.125, 4095},
		{0.25, 8191},
		{0.5, 16383},
		{1.0, 32767},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprint(test.arg), func(t *testing.T) {
			t.Parallel()
			actual := F32ToS16(test.arg)
			assert.Equal(t, test.expected, actual, "got %v, expected %v", actual, test.expected)
		})
	}
}

func TestS16toF32toS16RoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		arg int16
	}{
		{-32767},
		{-16383},
		{-8191},
		{-4095},
		{-2047},
		{-1023},
		{-511},
		{-255},
		{0},
		{255},
		{511},
		{1023},
		{2047},
		{4095},
		{8191},
		{16383},
		{32767},
	}
	for _, test := range testCases {
		t.Run(strconv.Itoa(int(test.arg)), func(t *testing.T) {
			t.Parallel()
			intermediate := S16ToF32(test.arg)
			result := F32ToS16(intermediate)
			assert.Equal(t, test.arg, result, "got %v, expected %v, intemediate %v", result, test.arg, intermediate)
		})
	}
}

func TestF32LEAdjustVolume(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		samples  []float32
		volume   float64
		expected []float32
	}{
		{
			name:     "volume 1.0 returns unchanged",
			samples:  []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
			volume:   1.0,
			expected: []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
		},
		{
			name:     "volume 0.5 halves amplitude",
			samples:  []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
			volume:   0.5,
			expected: []float32{-0.5, -0.25, 0.0, 0.25, 0.5},
		},
		{
			name:     "volume 0.0 silences audio",
			samples:  []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
			volume:   0.0,
			expected: []float32{0.0, 0.0, 0.0, 0.0, 0.0},
		},
		{
			name:     "volume 2.0 doubles amplitude but clamps",
			samples:  []float32{-0.75, -0.5, 0.0, 0.5, 0.75},
			volume:   2.0,
			expected: []float32{-1.0, -1.0, 0.0, 1.0, 1.0},
		},
		{
			name:     "volume > 1.0 clamps at limits",
			samples:  []float32{-1.0, 1.0},
			volume:   1.5,
			expected: []float32{-1.0, 1.0},
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := F32AdjustVolume(test.samples, test.volume)
			assert.Equal(t, test.expected, actual)
		})
	}
}
