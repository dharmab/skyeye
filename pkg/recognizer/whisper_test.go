package recognizer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/wav"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const dataDir = "testdata"

var modelPath = os.Getenv("SKYEYE_WHISPER_MODEL")

type sample struct {
	filename string
	pcm      []float32
}

func loadSample(b *testing.B, info fs.FileInfo) sample {
	b.Helper()
	_, err := fmt.Printf("Loading %s\n", info.Name())
	require.NoError(b, err)
	f, err := os.Open(filepath.Join(dataDir, info.Name()))
	require.NoError(b, err)
	streamer, format, err := wav.Decode(f)
	require.NoError(b, err)
	defer streamer.Close()
	assert.Equal(b, beep.SampleRate(16000), format.SampleRate)
	assert.Equalf(b, 1, format.NumChannels, "expected 1 channel, got %d", format.NumChannels)
	f32Data := make([]float32, 0)
	for {
		samples := make([][2]float64, 512)
		n, ok := streamer.Stream(samples)
		for i := range n {
			f64Sample := samples[i][0]
			f32Sample := float32(f64Sample / 1)
			f32Data = append(f32Data, f32Sample)
		}
		switch {
		case n == len(samples) && ok:
			continue
		case 0 < n && n < len(samples) && ok:
			return sample{info.Name(), f32Data}
		case !ok:
			require.NoError(b, streamer.Err())
		}
	}
}

func loadSamples(b *testing.B) []sample {
	b.Helper()
	entries, err := os.ReadDir(dataDir)
	require.NoError(b, err)
	var samples []sample
	for _, entry := range entries {
		info, err := entry.Info()
		require.NoError(b, err)
		endsWithWav := filepath.Ext(info.Name()) == ".wav"
		if !info.IsDir() && endsWithWav {
			sample := loadSample(b, info)
			samples = append(samples, sample)
		}
	}
	return samples
}

func BenchmarkWhisperRecognizer(b *testing.B) {
	samples := loadSamples(b)
	model, err := whisper.New(modelPath)
	require.NoError(b, err)
	recognizer := NewWhisperRecognizer(&model, "Thunderhead")
	b.ResetTimer()
	for _, sample := range samples {
		b.Run(sample.filename, func(b *testing.B) {
			for b.Loop() {
				_, err := recognizer.Recognize(b.Context(), sample.pcm, true)
				require.NoError(b, err)
			}
		})
	}
}
