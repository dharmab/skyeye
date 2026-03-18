// Package model provides download and verification of Pocket TTS model files.
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

// DirName is the subdirectory name used for the Pocket TTS model within a models directory.
const DirName = "pocket"

const modelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/tts-models/sherpa-onnx-pocket-tts-int8-2026-01-26.tar.bz2"

// archiveHash is the expected SHA256 hash of the downloaded tar.bz2 archive.
const archiveHash = "2f3b88823cbbb9bf0b2477ec8ae7b3fec417b3a87b6bb5f256dba66f2ad967cb"

// Model file names for Pocket TTS.
const (
	FilenameLmMain          = "lm_main.int8.onnx"
	FilenameLmFlow          = "lm_flow.int8.onnx"
	FilenameDecoder         = "decoder.int8.onnx"
	FilenameEncoder         = "encoder.onnx"
	FilenameTextConditioner = "text_conditioner.onnx"
	FilenameVocabJSON       = "vocab.json"
	FilenameTokenScoresJSON = "token_scores.json"
)

// Filenames lists the filenames required for the Pocket TTS model.
var Filenames = []string{
	FilenameLmMain,
	FilenameLmFlow,
	FilenameDecoder,
	FilenameEncoder,
	FilenameTextConditioner,
	FilenameVocabJSON,
	FilenameTokenScoresJSON,
}

// fileHashes maps each model filename to its expected SHA256 hash.
var fileHashes = map[string]string{ //nolint:gosec // SHA256 hashes for model verification, not credentials
	FilenameLmMain:          "bfc0c7e7e3d72864fa3bb2ee499f62f21ddc1474b885f5f3ca570f8be73e787e",
	FilenameLmFlow:          "8d627d235c44a597da908e1085ebe241cbbe358964c502c5a5063d18851a5529",
	FilenameDecoder:         "12b0857402d31aead94df19d6783b4350d1f740e811f3a3202c70ad89ae11eea",
	FilenameEncoder:         "e8f2f6d301ffb96e398b138a7dc6d3038622d236044636b73d920bab85890260",
	FilenameTextConditioner: "0b84e837d7bfaf2c896627b03e3f080320309f37f4fc7df7698c644f7ba5e6b1",
	FilenameVocabJSON:       "6fb646346cf931016f70c4921aab0900ce7a304b893cb02135c74e294abfea01",
	FilenameTokenScoresJSON: "5be2f278caf9b9800741f0fd82bff677f4943ec764c356f907213434b622d958",
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

// Download downloads the Pocket TTS model archive, verifies its SHA256 hash,
// extracts the required files into dir, and verifies their individual hashes.
func Download(ctx context.Context, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	log.Info().Str("url", modelURL).Msg("downloading Pocket TTS model")

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

	tmpFile, err := os.CreateTemp("", "pocket-model-*.tar.bz2")
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
