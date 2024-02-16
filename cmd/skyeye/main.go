package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dharmab/skyeye/internal/application"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/lithammer/shortuuid"
)

func main() {
	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stderr,
				&slog.HandlerOptions{
					Level: slog.LevelDebug,
				},
			),
		),
	)

	// CGO_ENABLED=1 LIBRARY_PATH=$(pwd)/../whisper.cpp C_INCLUDE_PATH=$(pwd)/../third_party/whisper.cpp go build ./cmd/skyeye/

	DCSGRPCAddress := flag.String("dcs-grpc-server-address", "localhost:50051", "address of the DCS-gRPC server")
	GRPCConnectionTimeout := flag.Duration("grpc-connection-timeout", 2*time.Second, "gRPC connection timeout")
	SRSAddress := flag.String("srs-server-address", "localhost:5002", "address of the SRS server")
	SRSConnectionTimeout := flag.Duration("srs-connection-timeout", 10*time.Second, "")
	SRSClientName := flag.String("srs-client-name", "SkyEye Bot", "SRS client name. Appears in the client list and in in-game transmissions")
	SRSExternalAWACSModePassword := flag.String("srs-eam-password", "", "SRS external AWACS mode password")
	SRSFrequency := flag.Float64("srs-frequency", 133.0, "AWACS frequency")
	SRSCoalition := flag.String("srs-coalition", "blue", "SRS Coalition (either blue or red)")
	WhisperModelPath := flag.String("whisper-model", "", "Path to whisper.cpp model")

	slog.Info("parsing CLI flags")
	flag.Parse()

	var coalition srs.Coalition
	if strings.EqualFold(*SRSCoalition, "blue") {
		coalition = srs.CoalitionBlue
	} else if strings.EqualFold(*SRSCoalition, "red") {
		coalition = srs.CoalitionRed
	} else {
		exitOnErr(errors.New("srs-coalition must be either blue or red"))
	}

	slog.Info("loading whisper model", "path", *WhisperModelPath)
	whisperModel, err := whisper.New(*WhisperModelPath)
	if err != nil {
		exitOnErr(fmt.Errorf("failed to load whisper model: %w", err))
	}
	defer whisperModel.Close()

	slog.Info("generating client GUID")
	clientGUID := shortuuid.New()

	config := application.Configuration{
		DCSGRPCAddress:               *DCSGRPCAddress,
		GRPCConnectionTimeout:        *GRPCConnectionTimeout,
		SRSAddress:                   *SRSAddress,
		SRSConnectionTimeout:         *SRSConnectionTimeout,
		SRSClientName:                *SRSClientName,
		SRSClientGUID:                clientGUID,
		SRSExternalAWACSModePassword: *SRSExternalAWACSModePassword,
		SRSFrequency:                 *SRSFrequency,
		SRSCoalition:                 coalition,
		WhisperModel:                 whisperModel,
	}

	ctx := context.Background()
	slog.Info("starting application")
	app, err := application.NewApplication(ctx, config)
	exitOnErr(err)
	err = app.Run(ctx)
	exitOnErr(err)
}

func exitOnErr(err error) {
	if err != nil {
		slog.With("error", err).Error("application exiting with error")
		os.Exit(1)
	}
}
