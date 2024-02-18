// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/simpleradio/audio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"

	"github.com/ebitengine/oto/v3"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Configuration for the SkyEye application.
type Configuration struct {
	// DCSGRPCAddress is the network address of the DCS-gRPC server (including port)
	DCSGRPCAddress string
	// GRPCConnectionTimeout is the connection timeout for connecting to DCS-gRPC
	GRPCConnectionTimeout time.Duration
	// SRSAddress is the network address of the SimpleRadio Standalone server (including port)
	SRSAddress string
	// SRSConnectionTimeout is the connection timeout for connecting to the SimpleRadio Standalone server
	SRSConnectionTimeout time.Duration
	// SRSClientName is the name of the bot that will appear in the client list and in in-game transmissions
	SRSClientName string
	// SRSExternalAWACSModePassword is the password for connecting to the SimpleRadio Standalone server using External AWACS Mode
	SRSExternalAWACSModePassword string
	// SRSFrequency is the radio frequency the bot will listen to and talk on in Hz
	SRSFrequency float64
	// SRSCoalition is the coalition that the bot will act on
	SRSCoalition srs.Coalition
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel whisper.Model
}

// Application is the interface for running the SkyEye application.
type Application interface {
	// Run runs the SkyEye application. It should be called exactly once.
	Run(context.Context) error
}

// app implements the Application.
type app struct {
	// dcsClient is a DCS-gRPC client
	dcsClient dcs.DCSClient
	// srsClient is a SimpleRadio Standalone client
	srsClient simpleradio.Client
	// whisper is a whisper.cpp model used for Speech To Text
	whisper whisper.Model
	// otoCtx is an oto context used for playing audio. This is only used for debugging purposes.
	// I should remove this when the SRS integration is stabilized.
	otoCtx oto.Context
}

// NewApplication constructs a new Application.
func NewApplication(ctx context.Context, config Configuration) (Application, error) {
	slog.Info("constructing DCS client")
	dcsClient, err := dcs.NewDCSClient(
		ctx,
		dcs.ClientConfiguration{
			Address:           config.DCSGRPCAddress,
			ConnectionTimeout: config.GRPCConnectionTimeout,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructing SRS client")
	srsClient, err := simpleradio.NewClient(
		srs.ClientConfiguration{
			Address:                   config.SRSAddress,
			ConnectionTimeout:         config.SRSConnectionTimeout,
			ClientName:                config.SRSClientName,
			ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
			Coalition:                 config.SRSCoalition,
			Radio: srs.Radio{
				Frequency:  config.SRSFrequency,
				Modulation: srs.ModulationAM,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	app := &app{
		dcsClient: dcsClient,
		srsClient: srsClient,
		whisper:   config.WhisperModel,
	}
	return app, nil
}

// Run implements Application.Run.
func (a *app) Run(ctx context.Context) error {
	defer func() {
		slog.Info("closing connection to DCS-gRPC server")
		err := a.dcsClient.Close()
		if err != nil {
			slog.Error("failed to close connection to DCS-gRPC server", "error", err)
		}
	}()

	go func() {
		slog.Info("running SRS client")
		if err := a.srsClient.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("error running SRS client", "error", err)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case sample := <-a.srsClient.Receive():
			slog.Debug("receiving sample from SRS client")
			text, err := a.recognizeAudio(sample)
			if err != nil {
				slog.Error("error recongizing audio sample", "error", err)
			} else {
				slog.Info("recognized audio", "text", text)
			}
		}
	}
}

// recognizeAudio recognizes audio using the whisper model. This needs to be moved into a separate package...
func (a *app) recognizeAudio(sample audio.Audio) (string, error) {
	wCtx, err := a.whisper.NewContext()
	if err != nil {
		return "", fmt.Errorf("error creating new speech-to-text context: %w", err)
	}
	slog.Info("processing sample")
	err = wCtx.Process(sample, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error processing audio sample: %w", err)
	}

	var textBuilder strings.Builder
	for {
		segment, err := wCtx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		slog.Info(
			"processed segment",
			"start", segment.Start,
			"end", segment.End,
			"text", segment.Text,
		)
		textBuilder.WriteString(fmt.Sprintf("%s\n", segment.Text))
	}
	return textBuilder.String(), nil
}
