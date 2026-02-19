// Package parakeet provides speech recognition using the NVIDIA Parakeet TDT model via sherpa-onnx.
package parakeet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dharmab/skyeye/pkg/recognizer"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/rs/zerolog/log"
)

// DirName is the subdirectory name used for the Parakeet model within a models directory.
const DirName = "parakeet"

// filenames lists the filenames required for the Parakeet TDT model.
var filenames = []string{
	"encoder.int8.onnx",
	"decoder.int8.onnx",
	"joiner.int8.onnx",
	"tokens.txt",
}

// fileHashes maps each model filename to its expected SHA256 hash.
var fileHashes = map[string]string{
	"encoder.int8.onnx": "a32b12d17bbbc309d0686fbbcc2987b5e9b8333a7da83fa6b089f0a2acd651ab",
	"decoder.int8.onnx": "b6bb64963457237b900e496ee9994b59294526439fbcc1fecf705b31a15c6b4e",
	"joiner.int8.onnx":  "7946164367946e7f9f29a122407c3252b680dbae9a51343eb2488d057c3c43d2",
	"tokens.txt":        "ec182b70dd42113aff6c5372c75cac58c952443eb22322f57bbd7f53977d497d",
}

// FileNotFoundError indicates that a required model file is missing.
type FileNotFoundError struct {
	Path string
	Err  error
}

func (e *FileNotFoundError) Error() string {
	return "model file not found: " + e.Path
}

func (e *FileNotFoundError) Unwrap() error {
	return e.Err
}

// CorruptFileError indicates that a model file exists but has an incorrect hash.
type CorruptFileError struct {
	Path     string
	Expected string
	Actual   string
}

func (e *CorruptFileError) Error() string {
	return fmt.Sprintf("model file %s: hash mismatch (expected %s, got %s)", e.Path, e.Expected, e.Actual)
}

// Verify checks that all model files exist in dir and match their expected SHA256 hashes.
// All files are checked and all errors are collected into a single joined error.
func Verify(dir string) error {
	var errs []error
	for _, name := range filenames {
		if err := verifyFile(filepath.Join(dir, name)); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func verifyFile(fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return &FileNotFoundError{Path: fpath, Err: err}
		}
		return fmt.Errorf("opening model file %s: %w", fpath, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("reading model file %s: %w", fpath, err)
	}
	actual := hex.EncodeToString(h.Sum(nil))
	basename := filepath.Base(fpath)
	expected := fileHashes[basename]
	if actual != expected {
		return &CorruptFileError{Path: fpath, Expected: expected, Actual: actual}
	}
	return nil
}

type parakeetRecognizer struct {
	recognizer *sherpa.OfflineRecognizer
}

var _ recognizer.Recognizer = &parakeetRecognizer{}

// NewRecognizer creates a new recognizer using NVIDIA Parakeet TDT via sherpa-onnx.
// modelDir must contain the model files (encoder, decoder, joiner ONNX files and tokens.txt).
func NewRecognizer(modelDir string) (recognizer.Recognizer, error) {
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
		return nil, errors.New("failed to create offline recognizer from model files")
	}

	return &parakeetRecognizer{recognizer: rec}, nil
}

// Close releases the recognizer resources.
func (r *parakeetRecognizer) Close() error {
	sherpa.DeleteOfflineRecognizer(r.recognizer)
	return nil
}

// Recognize implements [recognizer.Recognizer] using NVIDIA Parakeet TDT via sherpa-onnx.
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
