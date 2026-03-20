package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// Setup verifies a model's files and downloads them if missing.
// Corrupt files always return an error. Missing files are passed to the
// download function; if download is nil, missing files return an error.
func Setup(ctx context.Context, name, dir string, verify func(string) error, download func(context.Context, string) error) error {
	logger := log.With().Str("model", name).Str("directory", dir).Logger()

	logger.Info().Msg("verifying model files")
	err := verify(dir)
	if err == nil {
		logger.Info().Msg("model files verified")
		return nil
	}

	// Corrupt files should not be silently re-downloaded.
	if _, ok := errors.AsType[*CorruptFileError](err); ok {
		return fmt.Errorf("model %s files on disk failed verification: %w", name, err)
	}

	// Missing files can be downloaded.
	if _, ok := errors.AsType[*FileNotFoundError](err); ok {
		if download == nil {
			return fmt.Errorf("model %s files not found: %w", name, err)
		}
		logger.Warn().Err(err).Msg("model files not found")
		logger.Info().Msg("downloading model files")
		if dlErr := download(ctx, dir); dlErr != nil {
			return fmt.Errorf("failed to download model %s: %w", name, dlErr)
		}
		return nil
	}

	// Unexpected error (e.g. permission denied).
	return fmt.Errorf("failed to verify model %s files: %w", name, err)
}
