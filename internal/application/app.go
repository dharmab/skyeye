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

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

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
	// SRSClientGUID is a client-provided GUID which the server uses to distinguish clients
	SRSClientGUID string
	// SRSExternalAWACSModePassword is the password for connecting to the SimpleRadio Standalone server using External AWACS Mode
	SRSExternalAWACSModePassword string
	// SRSFrequency is the radio frequency the bot will listen to and talk on in MHz
	SRSFrequency float64
	// SRSCoalition is the coalition that the bot will act on
	SRSCoalition srs.Coalition
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel whisper.Model
}

type Application interface {
	Run(context.Context) error
}

type app struct {
	dcsClient dcs.DCSClient
	srsClient simpleradio.Client
	whisper   whisper.Model
}

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
			GUID:                      config.SRSClientGUID,
			Address:                   config.SRSAddress,
			ConnectionTimeout:         config.SRSConnectionTimeout,
			ClientName:                config.SRSClientName,
			ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
			Coalition:                 config.SRSCoalition,
			Frequency: srs.Frequency{
				Frequency:  config.SRSFrequency,
				Modulation: srs.ModulationAM,
				Encryption: 0,
			},
		},
		srs.RadioInfo{
			Radios: []srs.Radio{{
				Frequency:  config.SRSFrequency,
				Modulation: srs.ModulationAM,
			}},
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
			a.recognizeAudio(sample)
		}
	}
}

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
