// Package parakeet provides speech recognition using Nvidia's Parakeet model.
package parakeet

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/dharmab/skyeye/pkg/pcm/rate"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/rs/zerolog/log"
)

type parakeetRecognizer struct {
	recognizer *sherpa.OfflineRecognizer
}

var _ recognizer.Recognizer = &parakeetRecognizer{}

// NewRecognizer creates a new recognizer using NVIDIA Parakeet. modelDir is a
// directory containing contain the model files (i.e the ONNX runtime files and
// tokens.txt).
func NewRecognizer(modelDir string) (recognizer.Recognizer, error) {
	config := sherpa.OfflineRecognizerConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: int(rate.Wideband.Hertz()),
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OfflineModelConfig{
			Transducer: sherpa.OfflineTransducerModelConfig{
				Encoder: filepath.Join(modelDir, model.Filenames[0]),
				Decoder: filepath.Join(modelDir, model.Filenames[1]),
				Joiner:  filepath.Join(modelDir, model.Filenames[2]),
			},
			Tokens:    filepath.Join(modelDir, model.Filenames[3]),
			ModelType: "nemo_transducer",
		},
		DecodingMethod: "greedy_search",
	}

	rec := sherpa.NewOfflineRecognizer(&config)
	if rec == nil {
		return nil, errors.New("failed to create offline recognizer from model files")
	}

	return &parakeetRecognizer{recognizer: rec}, nil
}

// Close releases the recognizer resources.
func (r *parakeetRecognizer) Close() error {
	sherpa.DeleteOfflineRecognizer(r.recognizer)
	return nil
}

// Recognize implements [recognizer.Recognizer] using NVIDIA Parakeet.
func (r *parakeetRecognizer) Recognize(_ context.Context, pcm []float32, enableTranscriptionLogging bool) (string, error) {
	stream := sherpa.NewOfflineStream(r.recognizer)
	if stream == nil {
		return "", errors.New("failed to create offline stream")
	}
	defer sherpa.DeleteOfflineStream(stream)

	stream.AcceptWaveform(int(rate.Wideband.Hertz()), pcm)
	r.recognizer.Decode(stream)
	result := stream.GetResult()
	if result == nil {
		return "", errors.New("recognition returned no result")
	}

	text := strings.TrimSpace(result.Text)

	event := log.Debug()
	if enableTranscriptionLogging {
		event = event.Str("text", text)
	}
	event.Msg("recognition complete")

	return text, nil
}
