package recognizer

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/rs/zerolog/log"
)

//go:embed model
var modelFS embed.FS

type parakeetRecognizer struct {
	recognizer *sherpa.OfflineRecognizer
	modelDir   string
}

var _ Recognizer = &parakeetRecognizer{}

// NewParakeetRecognizer creates a new recognizer using NVIDIA Parakeet TDT via sherpa-onnx.
// Model files are embedded in the binary and extracted to a temporary directory.
func NewParakeetRecognizer() (Recognizer, error) {
	modelDir, err := extractModel()
	if err != nil {
		return nil, fmt.Errorf("failed to extract embedded model files: %w", err)
	}

	config := sherpa.OfflineRecognizerConfig{
		FeatConfig: sherpa.FeatureConfig{
			SampleRate: 16000,
			FeatureDim: 80,
		},
		ModelConfig: sherpa.OfflineModelConfig{
			Transducer: sherpa.OfflineTransducerModelConfig{
				Encoder: filepath.Join(modelDir, "encoder.int8.onnx"),
				Decoder: filepath.Join(modelDir, "decoder.int8.onnx"),
				Joiner:  filepath.Join(modelDir, "joiner.int8.onnx"),
			},
			Tokens:    filepath.Join(modelDir, "tokens.txt"),
			ModelType: "nemo_transducer",
		},
		DecodingMethod: "greedy_search",
	}

	rec := sherpa.NewOfflineRecognizer(&config)
	if rec == nil {
		return nil, errors.New("failed to create offline recognizer from extracted model files")
	}

	return &parakeetRecognizer{recognizer: rec, modelDir: modelDir}, nil
}

// extractModel writes embedded model files to a temporary directory and returns its path.
func extractModel() (string, error) {
	dir, err := os.MkdirTemp("", "skyeye-parakeet-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	entries, err := modelFS.ReadDir("model")
	if err != nil {
		return "", fmt.Errorf("failed to read embedded model directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := modelFS.ReadFile(filepath.Join("model", entry.Name()))
		if err != nil {
			return "", fmt.Errorf("failed to read embedded file %s: %w", entry.Name(), err)
		}
		dst := filepath.Join(dir, entry.Name())
		if err := os.WriteFile(dst, data, 0o600); err != nil {
			return "", fmt.Errorf("failed to write model file %s: %w", entry.Name(), err)
		}
		log.Debug().Str("file", entry.Name()).Str("dir", dir).Msg("extracted model file")
	}

	return dir, nil
}

// Close cleans up the extracted model files.
func (r *parakeetRecognizer) Close() error {
	sherpa.DeleteOfflineRecognizer(r.recognizer)
	if r.modelDir != "" {
		return os.RemoveAll(r.modelDir)
	}
	return nil
}

// Recognize implements [Recognizer.Recognize] using NVIDIA Parakeet TDT via sherpa-onnx.
func (r *parakeetRecognizer) Recognize(_ context.Context, pcm []float32, enableTranscriptionLogging bool) (string, error) {
	stream := sherpa.NewOfflineStream(r.recognizer)
	if stream == nil {
		return "", errors.New("failed to create offline stream")
	}
	defer sherpa.DeleteOfflineStream(stream)

	stream.AcceptWaveform(16000, pcm)
	r.recognizer.Decode(stream)
	result := stream.GetResult()

	text := strings.TrimSpace(result.Text)

	event := log.Debug()
	if enableTranscriptionLogging {
		event = event.Str("text", text)
	}
	event.Msg("recognition complete")

	return text, nil
}
