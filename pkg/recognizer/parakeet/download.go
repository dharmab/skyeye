package parakeet

import (
	"archive/tar"
	"compress/bzip2"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

const modelURL = "https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-nemo-parakeet-tdt-0.6b-v2-int8.tar.bz2"

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

	needed := make(map[string]bool, len(filenames))
	for _, f := range filenames {
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

	if extracted != len(filenames) {
		return fmt.Errorf("expected %d model files in archive, found %d", len(filenames), extracted)
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
