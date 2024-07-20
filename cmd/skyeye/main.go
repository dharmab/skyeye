package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dharmab/skyeye/internal/application"
	"github.com/dharmab/skyeye/internal/conf"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

func main() {
	// Set up an application-scoped context and a cancel function to shut down the application.
	ctx, cancel := context.WithCancel(context.Background())

	// Set up a signal handler to shut down the application when an interrupt or TERM signal is received.
	interuptChan := make(chan os.Signal, 1)
	signal.Notify(interuptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-interuptChan
		log.Info().Interface("signal", s).Msg("received shutdown signal")
		cancel()
		time.Sleep(time.Second)
		os.Exit(0)
	}()

	// Parse configuration from CLI flags.
	LogLevel := flag.String("log-level", "info", "logging level (trace, debug, info, warn, error, fatal)")
	LogFormat := flag.String("log-format", "json", "logging format (json, pretty)")
	DCSGRPCAddress := flag.String("dcs-grpc-server-address", "localhost:50051", "address of the DCS-gRPC server")
	GRPCConnectionTimeout := flag.Duration("grpc-connection-timeout", 2*time.Second, "gRPC connection timeout")
	SRSAddress := flag.String("srs-server-address", "127.0.0.1:5002", "address of the SRS server")
	SRSConnectionTimeout := flag.Duration("srs-connection-timeout", 10*time.Second, "")
	SRSClientName := flag.String("srs-client-name", "Skyeye", "SRS client name. Appears in the client list and in in-game transmissions")
	SRSExternalAWACSModePassword := flag.String("srs-eam-password", "", "SRS external AWACS mode password")
	SRSFrequency := flag.Float64("srs-frequency", 251000000.0, "AWACS frequency in Hertz")
	SRSCoalition := flag.String("srs-coalition", "blue", "SRS Coalition (either blue or red)")
	WhisperModelPath := flag.String("whisper-model", "", "Path to whisper.cpp model")

	flag.Parse()

	if strings.EqualFold(*LogFormat, "pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	var level zerolog.Level
	switch strings.ToLower(*LogLevel) {
	case "error":
		level = zerolog.ErrorLevel
	case "warn":
		level = zerolog.WarnLevel
	case "info":
		level = zerolog.InfoLevel
	case "debug":
		level = zerolog.DebugLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	var coalition srs.Coalition
	log.Info().Str("coalition", *SRSCoalition).Msg("setting GCI coalition")
	if strings.EqualFold(*SRSCoalition, "blue") {
		coalition = srs.CoalitionBlue
	} else if strings.EqualFold(*SRSCoalition, "red") {
		coalition = srs.CoalitionRed
	} else {
		exitOnErr(errors.New("srs-coalition must be either blue or red"))
	}
	log.Info().Int("id", int(coalition)).Msg("GCI coalition set")

	// Load whisper.cpp model
	log.Info().Str("path", *WhisperModelPath).Msg("loading whisper model")
	whisperModel, err := whisper.New(*WhisperModelPath)
	if err != nil {
		exitOnErr(fmt.Errorf("failed to load whisper model: %w", err))
	}
	log.Info().Msg("whisper model loaded")
	defer whisperModel.Close()

	// Configure and run the application.
	config := conf.Configuration{
		DCSGRPCAddress:               *DCSGRPCAddress,
		GRPCConnectionTimeout:        *GRPCConnectionTimeout,
		SRSAddress:                   *SRSAddress,
		SRSConnectionTimeout:         *SRSConnectionTimeout,
		SRSClientName:                *SRSClientName,
		SRSExternalAWACSModePassword: *SRSExternalAWACSModePassword,
		SRSFrequency:                 *SRSFrequency,
		SRSCoalition:                 coalition,
		WhisperModel:                 whisperModel,
	}

	log.Info().Msg("starting application")
	app, err := application.NewApplication(ctx, config)
	exitOnErr(err)
	err = app.Run(ctx)
	exitOnErr(err)
}

// exitOnErr logs the error and exits the application if the error is not nil.
func exitOnErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("application exiting with error")
		os.Exit(1)
	}
}
