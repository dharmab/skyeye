package telemetry

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// FileReader reads telemetry data from a file.
type FileReader struct {
	streamingClient
	filePath string
}

// NewFileClient creates a new telemetry client for reading data from a file.
func NewFileClient(
	filePath string,
	updateInterval time.Duration,
) *FileReader {
	return &FileReader{
		streamingClient: *newStreamingClient(updateInterval),
		filePath:        filePath,
	}
}

// Run reads telemetry data from the file.
func (r *FileReader) Run(ctx context.Context) error {
	f, err := openFile(r.filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	if err := r.handleLines(ctx, reader); err != nil {
		return fmt.Errorf("error reading data: %w", err)
	}
	return nil
}

func openFile(path string) (io.ReadCloser, error) {
	logger := log.With().Str("path", path).Logger()
	// ZIP archive
	if isZipped, err := isZipped(path); err != nil {
		return nil, err
	} else if isZipped {
		logger.Info().Msg("opening compressed ACMI file")
		reader, err := zip.OpenReader(path)
		if err != nil {
			return nil, err
		}
		logger.Info().Int("files", len(reader.File)).Msg("searching for ACMI file in ZIP archive")
		for _, f := range reader.File {
			if strings.HasSuffix(f.Name, ".txt.acmi") {
				logger.Info().Str("file", f.Name).Msg("found ACMI file in ZIP archive")
				acmi, err := f.Open()
				if err != nil {
					return nil, err
				}
				return acmi, nil
			}
		}
	}

	// Text file
	logger.Info().Msg("opening ACMI file")
	acmi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return acmi, nil
}

func isZipped(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil {
		return false, err
	}

	contentType := http.DetectContentType(buf)
	return contentType == "application/zip", nil
}
