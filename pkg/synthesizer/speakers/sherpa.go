package speakers

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/rs/zerolog/log"
)

//go:embed kokoro
var kokoroData embed.FS

type sherpaSynth struct {
	tts       *sherpa.OfflineTts
	dataDir   string
	speed     float32
	speakerID int
}

var _ Speaker = (*sherpaSynth)(nil)

const (
	kokoroModelFilename   = "model.onnx"
	kokoroVoicesFilename  = "voices.bin"
	kokoroTokensFilename  = "tokens.txt"
	espeakNGDirectoryName = "espeak-ng-data"
)

const (
	feminineSpeakerID  = 8
	masculineSpeakerID = 12
)

// NewSherpaSpeaker creates a Speaker powered by sherpa-onnx
// (https://k2-fsa.github.io/sherpa/onnx/index.html) and Kokoro
// (https://kokorotts.org/)
func NewSherpaSpeaker(v voices.Voice, playbackSpeed float32) (Speaker, error) {
	var speakerID int
	if v == voices.FeminineVoice {
		speakerID = feminineSpeakerID
	} else if v == voices.MasculineVoice {
		speakerID = masculineSpeakerID
	}

	// Unpack the Kokoro data into a temporary directory
	dataDir, err := os.MkdirTemp("", "sherpa")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory for Kokoro data: %w", err)
	}
	if err := unpack(kokoroData, dataDir); err != nil {
		if err := os.RemoveAll(dataDir); err != nil {
			log.Error().Err(err).Str("path", dataDir).Msg("failed to remove temporary directory")
		}
		return nil, fmt.Errorf("failed to unpack Kokoro data into temporary directory: %w", err)
	}

	cfg := &sherpa.OfflineTtsConfig{}
	cfg.Model.Kokoro = sherpa.OfflineTtsKokoroModelConfig{
		Model:   filepath.Join(dataDir, kokoroModelFilename),
		Voices:  filepath.Join(dataDir, kokoroVoicesFilename),
		Tokens:  filepath.Join(dataDir, kokoroTokensFilename),
		DataDir: filepath.Join(dataDir, espeakNGDirectoryName),
	}

	tts := sherpa.NewOfflineTts(cfg)

	synth := &sherpaSynth{
		tts:       tts,
		dataDir:   dataDir,
		speed:     playbackSpeed,
		speakerID: speakerID,
	}

	return synth, nil
}

// Unpack a filesystem into a destination directory.
func unpack(filesystem fs.FS, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := fs.WalkDir(filesystem, ".", func(sourcePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, sourcePath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		sourceFile, err := filesystem.Open(sourcePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sourceFile)
		return err
	}); err != nil {
		return fmt.Errorf("failed to unpack embedded files: %w", err)
	}
	return nil
}

// Say implements [Speaker.Say] using sherpa-onnx and Kokoro.
func (s *sherpaSynth) Say(text string) ([]float32, error) {
	audio := s.tts.Generate(text, s.speakerID, s.speed)
	log.Info().Int("sampleRate", audio.SampleRate).Str("text", text).Msg("generated audio")
	if audio.SampleRate != int(targetSampleRate.Hertz()) {
		log.Warn().Int("sampleRate", audio.SampleRate).Int("targetSampleRate", int(targetSampleRate.Hertz())).Msg("sample rate mismatch")
	}
	return audio.Samples, nil
}

// Close releases resources.
func (s *sherpaSynth) Close() (err error) {
	log.Info().Msg("cleaning up Sherpa/Kokoro data")
	if s.tts != nil {
		log.Info().Msg("cleaning up Sherpa TTS")
		sherpa.DeleteOfflineTts(s.tts)
	}
	if s.dataDir != "" {
		logger := log.With().Str("path", s.dataDir).Logger()
		logger.Info().Msg("cleaning up Kokoro data directory")
		if rmErr := os.RemoveAll(s.dataDir); rmErr != nil {
			logger.Error().Err(rmErr).Msg("failed to remove Kokoro data directory")
			err = errors.Join(err, rmErr)
		}
	}
	return
}
