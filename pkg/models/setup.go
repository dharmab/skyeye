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
	verifyErr := verify(dir)
	if verifyErr == nil {
		logger.Info().Msg("model files successfully verified")
		return nil
	}

	// Corrupt files should not be silently re-downloaded.
	if _, ok := errors.AsType[*CorruptFileError](verifyErr); ok {
		return fmt.Errorf("model files on disk failed verification: %w", verifyErr)
	}

	// Missing files can be downloaded.
	if _, ok := errors.AsType[*FileNotFoundError](verifyErr); ok {
		if download == nil {
			return fmt.Errorf("model files not found: %w", verifyErr)
		}
		logger.Warn().Err(verifyErr).Msg("model files not found")
		logger.Info().Msg("downloading model files")
		if downloadErr := download(ctx, dir); downloadErr != nil {
			return fmt.Errorf("failed to download model %s: %w", name, downloadErr)
		}
		return nil
	}

	// Unexpected error (e.g. permission denied).
	return fmt.Errorf("failed to verify model files: %w", verifyErr)
}
