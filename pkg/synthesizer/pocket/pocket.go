// Package pocket provides a text-to-speech speaker using Pocket TTS via sherpa-onnx.
package pocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
	"github.com/dharmab/skyeye/pkg/synthesizer/pocket/voice"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

type options struct {
	voiceFile  string
	numSteps   int
	numThreads int
}

// Option configures Speaker behavior.
type Option func(*options)

// WithVoiceFile sets a custom WAV file for voice cloning reference audio.
// The file must be 16-bit PCM mono WAV. If the file cannot be read,
// the embedded default voice is used instead.
func WithVoiceFile(path string) Option {
	return func(o *options) {
		o.voiceFile = path
	}
}

// WithNumSteps sets the number of inference steps (default 10).
func WithNumSteps(n int) Option {
	return func(o *options) {
		o.numSteps = n
	}
}

// WithThreads sets the number of threads for ONNX Runtime inference (default 2).
// 1 thread performs worse in benchmarks but may be useful when running SkyEye
// alongside other applications such as DCS World.
// More than 2 threads regressed in performance due to threading overhead.
func WithThreads(n int) Option {
	return func(o *options) {
		o.numThreads = n
	}
}

// Speaker implements speakers.Speaker using Pocket TTS.
type Speaker struct {
	tts       *sherpa.OfflineTts
	genConfig sherpa.GenerationConfig
}

var _ speakers.Speaker = (*Speaker)(nil)

// New creates a Speaker. modelDir must contain the Pocket TTS model files.
func New(modelDir string, opts ...Option) (*Speaker, error) {
	o := &options{
		numSteps:   10,
		numThreads: 2,
	}
	for _, opt := range opts {
		opt(o)
	}

	config := sherpa.OfflineTtsConfig{
		Model: sherpa.OfflineTtsModelConfig{
			Pocket: sherpa.OfflineTtsPocketModelConfig{
				LmMain:          filepath.Join(modelDir, model.FilenameLmMain),
				LmFlow:          filepath.Join(modelDir, model.FilenameLmFlow),
				Decoder:         filepath.Join(modelDir, model.FilenameDecoder),
				Encoder:         filepath.Join(modelDir, model.FilenameEncoder),
				TextConditioner: filepath.Join(modelDir, model.FilenameTextConditioner),
				VocabJson:       filepath.Join(modelDir, model.FilenameVocabJSON),
				TokenScoresJson: filepath.Join(modelDir, model.FilenameTokenScoresJSON),
			},
			NumThreads: o.numThreads,
			Debug:      0,
		},
	}

	tts := sherpa.NewOfflineTts(&config)
	if tts == nil {
		return nil, errors.New("failed to create Pocket TTS from model files")
	}

	refAudio, refRate, err := loadReferenceAudio(o.voiceFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load reference audio: %w", err)
	}

	genConfig := sherpa.GenerationConfig{
		ReferenceAudio:      refAudio,
		ReferenceSampleRate: refRate,
		NumSteps:            o.numSteps,
		Speed:               1.0,
		Extra:               json.RawMessage(`{"temperature": 0.4}`),
	}

	return &Speaker{
		tts:       tts,
		genConfig: genConfig,
	}, nil
}

// Say implements speakers.Speaker.
func (s *Speaker) Say(ctx context.Context, text string) ([]float32, error) {
	cb := func(_ []float32, _ float32) bool {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	audio := s.tts.GenerateWithConfig(text, &s.genConfig, cb)
	if audio == nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("pocket TTS generation returned nil")
	}

	sourceRate := unit.Frequency(audio.SampleRate) * unit.Hertz
	resampled, err := downsample(audio.Samples, sourceRate)
	if err != nil {
		return nil, fmt.Errorf("failed to resample pocket TTS output: %w", err)
	}

	return resampled, nil
}

// Close releases C resources held by the TTS engine.
func (s *Speaker) Close() {
	sherpa.DeleteOfflineTts(s.tts)
}

func loadReferenceAudio(voiceFile string) ([]float32, int, error) {
	if voiceFile != "" {
		samples, sampleRate, err := loadVoiceFile(voiceFile)
		if err == nil {
			return samples, sampleRate, nil
		}
		log.Warn().Err(err).Str("path", voiceFile).Msg("failed to load voice file, falling back to default")
	}

	samples, sampleRate, err := voice.DecodeWAV(voice.DefaultVoice)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode embedded default voice: %w", err)
	}
	log.Info().Msg("using default reference voice")
	return samples, sampleRate, nil
}

func loadVoiceFile(path string) ([]float32, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read voice file: %w", err)
	}
	samples, sampleRate, err := voice.DecodeWAV(data)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode voice file: %w", err)
	}
	log.Info().Str("path", path).Msg("using custom reference voice")
	return samples, sampleRate, nil
}
