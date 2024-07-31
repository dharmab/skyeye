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
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/synthesizer"
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
	ACMIFile := flag.String("acmi-file", "", "path to ACMI file")
	TelemetryAddress := flag.String("telemetry-address", "127.0.0.1:42674", "address of the real-time telemetry service")
	TelemetryConnectionTimeout := flag.Duration("telemetry-connection-timeout", 10*time.Second, "")
	TelemetryPassword := flag.String("telemetry-password", "", "password for the real-time telemetry service")
	SRSAddress := flag.String("srs-server-address", "127.0.0.1:5002", "address of the SRS server")
	SRSConnectionTimeout := flag.Duration("srs-connection-timeout", 10*time.Second, "")
	SRSExternalAWACSModePassword := flag.String("srs-eam-password", "", "SRS external AWACS mode password")
	SRSFrequency := flag.Float64("srs-frequency", 251000000.0, "AWACS frequency in Hertz")
	GCICallsign := flag.String("callsign", "Magic", "GCI callsign. Used in radio transmissions")
	Coalition := flag.String("coalition", "blue", "Coalition (either blue or red)")
	RadarSweepInterval := flag.Duration("radar-sweep-interval", 2*time.Second, "Radar update tick rate")
	WhisperModelPath := flag.String("whisper-model", "", "Path to whisper.cpp model")
	Voice := flag.String("voice", "feminine", "Voice to use for SRS transmissions (feminine, masculine)")

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

	if *ACMIFile == "" && *TelemetryAddress == "" {
		exitOnErr(errors.New("either ACMI file or telemetry address must be provided"))
	}

	var coalition coalitions.Coalition
	log.Info().Str("coalition", *Coalition).Msg("setting GCI coalition")
	if strings.EqualFold(*Coalition, "blue") {
		coalition = coalitions.Blue
	} else if strings.EqualFold(*Coalition, "red") {
		coalition = coalitions.Red
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

	var voice synthesizer.Voice
	switch strings.ToLower(*Voice) {
	case "feminine":
		voice = synthesizer.FeminineVoice
	case "masculine":
		voice = synthesizer.MasculineVoice
	default:
		err = fmt.Errorf("unknown voice: %s", *Voice)
		exitOnErr(err)
	}

	// Configure and run the application.
	config := conf.Configuration{
		ACMIFile:                     *ACMIFile,
		TelemetryAddress:             *TelemetryAddress,
		TelemetryConnectionTimeout:   *TelemetryConnectionTimeout,
		TelemetryClientName:          *GCICallsign,
		TelemetryPassword:            *TelemetryPassword,
		SRSAddress:                   *SRSAddress,
		SRSConnectionTimeout:         *SRSConnectionTimeout,
		SRSClientName:                fmt.Sprintf("GCI %s [BOT]", *GCICallsign),
		SRSExternalAWACSModePassword: *SRSExternalAWACSModePassword,
		SRSFrequency:                 *SRSFrequency,
		Callsign:                     *GCICallsign,
		Coalition:                    coalition,
		RadarSweepInterval:           *RadarSweepInterval,
		WhisperModel:                 whisperModel,
		Voice:                        voice,
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
