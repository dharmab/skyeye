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

// DirName is the subdirectory name used for the Parakeet model within a models directory.
const DirName = "parakeet"

const modelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8.tar.bz2"

// Filenames lists the filenames required for the Parakeet TDT model.
var Filenames = []string{
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

// Download downloads the Parakeet TDT model archive, extracts the required
// files into dir, and verifies their SHA256 hashes.
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

	needed := make(map[string]bool, len(Filenames))
	for _, f := range Filenames {
		needed[f] = true
	}

	bzReader := bzip2.NewReader(resp.Body)
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

		// The archive contains files under a top-level directory; extract only the base name.
		base := filepath.Base(header.Name)
		if !needed[base] {
			continue
		}

		// Guard against path traversal.
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

	log.Info().Msg("verifying model file hashes")
	if err := Verify(dir); err != nil {
		return fmt.Errorf("model verification after download failed: %w", err)
	}

	log.Info().Msg("model download complete")
	return nil
}

func extractTarEntry(dst string, r io.Reader) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", dst, err)
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil { //nolint:gosec // archive source is trusted
		return fmt.Errorf("writing file %s: %w", dst, err)
	}
	return nil
}
