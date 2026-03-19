package voice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeWAV_DefaultVoice(t *testing.T) {
	t.Parallel()
	samples, sampleRate, err := DecodeWAV(DefaultVoice)
	require.NoError(t, err)
	assert.Greater(t, sampleRate, 0)
	assert.NotEmpty(t, samples)

	// Verify samples are in valid range [-1, 1]
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("sample %d out of range: %f", i, s)
			break
		}
	}
}

func TestDecodeWAV_InvalidData(t *testing.T) {
	t.Parallel()
	_, _, err := DecodeWAV([]byte("not a wav file"))
	assert.Error(t, err)
}

func TestDecodeWAV_TooShort(t *testing.T) {
	t.Parallel()
	_, _, err := DecodeWAV([]byte{0, 1, 2})
	assert.Error(t, err)
}
