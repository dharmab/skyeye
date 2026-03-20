// Package model provides download and verification of Parakeet TDT model files.
// This package has no CGO dependencies and can be built with CGO_ENABLED=0.
package model

import (
	"context"

	"github.com/dharmab/skyeye/pkg/models"
)

// Verify checks that all model files exist in dir and match their expected SHA256 hashes.
func Verify(dir string) error {
	return models.Verify(dir, Filenames, fileHashes)
}

// Download downloads the Parakeet TDT model archive, verifies its SHA256 hash,
// extracts the required files into dir, and verifies their individual hashes.
func Download(ctx context.Context, dir string) error {
	return models.Download(ctx, "parakeet", modelURL, archiveHash, dir, Filenames, fileHashes)
}
