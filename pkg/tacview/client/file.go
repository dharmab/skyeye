package client

import (
	"archive/zip"
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type fileClient struct {
	file io.ReadCloser
	*tacviewClient
}

func NewFileClient(path string, coalition coalitions.Coalition, updates chan<- sim.Updated, fades chan<- sim.Faded, updateInterval time.Duration) (Client, error) {
	f, err := openFile(path)
	if err != nil {
		return nil, err
	}
	return &fileClient{
		file: f,
		tacviewClient: &tacviewClient{
			coalition:      coalition,
			updates:        updates,
			fades:          fades,
			updateInterval: updateInterval,
			missionTime:    time.Unix(0, 0),
		},
	}, nil
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

func (c *fileClient) Run(ctx context.Context) error {
	reader := bufio.NewReader(c.file)
	acmi := acmi.New(c.coalition, reader, c.updateInterval)
	return c.tacviewClient.run(ctx, acmi)
}

func (c *fileClient) Bullseye() orb.Point {
	return c.tacviewClient.bullseye
}

func (c *fileClient) Time() time.Time {
	return c.tacviewClient.missionTime
}

func (c *fileClient) Close() error {
	return c.file.Close()
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
