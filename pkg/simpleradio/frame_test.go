package simpleradio

import (
	"math"
	"testing"

	"github.com/dharmab/skyeye/pkg/pcm/rate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/hraban/opus.v2"
)

// testTone generates a frame of audio loud enough to survive a lossy codec round trip.
func testTone(samples int) []float32 {
	tone := make([]float32, samples)
	for i := range tone {
		tone[i] = 0.5 * float32(math.Sin(2*math.Pi*440*float64(i)/rate.Wideband.Hertz()))
	}
	return tone
}

func TestFrameRoundTrip(t *testing.T) {
	t.Parallel()
	encoder, err := opus.NewEncoder(int(rate.Wideband.Hertz()), channels, opusApplicationVoIP)
	require.NoError(t, err)
	decoder, err := opus.NewDecoder(int(rate.Wideband.Hertz()), channels)
	require.NoError(t, err)

	original := testTone(int(frameSize))
	encoded, err := encodeFrame(encoder, original)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)
	assert.Less(t, len(encoded), encodingBufferSize)

	decoded, err := decodeFrame(decoder, encoded)
	require.NoError(t, err)
	assert.Len(t, decoded, int(frameSize), "a decoded frame is always a whole 40ms frame")

	// Opus is lossy, so compare energy rather than samples.
	var energy float64
	for _, sample := range decoded {
		energy += float64(sample) * float64(sample)
	}
	assert.Positive(t, energy, "decoded frame should carry the tone, not silence")
}

func TestDecodeFrameRejectsMalformedAudio(t *testing.T) {
	t.Parallel()
	decoder, err := opus.NewDecoder(int(rate.Wideband.Hertz()), channels)
	require.NoError(t, err)

	tests := []struct {
		name  string
		audio []byte
	}{
		{name: "empty", audio: []byte{}},
		{name: "garbage", audio: []byte{0xDE, 0xAD, 0xBE, 0xEF}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, decodeErr := decodeFrame(decoder, test.audio)
			assert.Error(t, decodeErr)
		})
	}
}

// frameSize is what ties the 40ms SRS frame to the wideband sample rate. If it drifts, every
// transmission's duration accounting drifts with it.
func TestFrameSize(t *testing.T) {
	t.Parallel()
	assert.Equal(t, int64(640), frameSize, "40ms of 16kHz mono audio")
	assert.Equal(t, frameSize, frameLength.Milliseconds()*int64(rate.Wideband.Kilohertz()))
}
