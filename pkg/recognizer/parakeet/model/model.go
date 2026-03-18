// Package model provides download and verification of Parakeet TDT model files.
// This package has no CGO dependencies and can be built with CGO_ENABLED=0.
package model

import (
	"archive/tar"
	"compress/bzip2"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

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
	for _, name := range Filenames {
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

// Download downloads the Parakeet TDT model archive, verifies its SHA256 hash,
// extracts the required files into dir, and verifies their individual hashes.
func Download(ctx context.Context, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	log.Info().Str("url", modelURL).Msg("downloading Parakeet TDT model")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download model: HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "parakeet-model-*.tar.bz2")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	h := sha256.New()
	if _, err := io.Copy(tmpFile, io.TeeReader(resp.Body, h)); err != nil {
		return fmt.Errorf("downloading archive: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != archiveHash {
		return fmt.Errorf("archive hash mismatch: expected %s, got %s", archiveHash, actual)
	}
	log.Info().Msg("archive hash verified")

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking temp file: %w", err)
	}

	if err := extractArchive(tmpFile, dir); err != nil {
		return err
	}

	log.Info().Msg("verifying model file hashes")
	if err := Verify(dir); err != nil {
		return fmt.Errorf("model verification after download failed: %w", err)
	}

	log.Info().Msg("model download complete")
	return nil
}

func extractArchive(r io.Reader, dir string) error {
	needed := make(map[string]bool, len(Filenames))
	for _, f := range Filenames {
		needed[f] = true
	}

	bzReader := bzip2.NewReader(r)
	tarReader := tar.NewReader(bzReader)

	extracted := 0
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar archive: %w", err)
		}

		base := filepath.Base(header.Name)
		if !needed[base] {
			continue
		}

		if strings.Contains(base, "..") {
			continue
		}

		dst := filepath.Join(dir, base)
		if err := extractTarEntry(dst, tarReader); err != nil {
			return err
		}
		log.Info().Str("file", base).Msg("extracted model file")
		extracted++
	}

	if extracted != len(Filenames) {
		return fmt.Errorf("expected %d model files in archive, found %d", len(Filenames), extracted)
	}
	return nil
}

func extractTarEntry(dst string, r io.Reader) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", dst, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil { //nolint:gosec // archive hash verified before extraction
		return fmt.Errorf("writing file %s: %w", dst, err)
	}
	return nil
}
