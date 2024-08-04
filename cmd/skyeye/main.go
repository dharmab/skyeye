package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/dharmab/skyeye/internal/application"
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
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
		log.Info().Any("signal", s).Msg("received shutdown signal")
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
	SRSFrequency := flag.Float64("srs-frequency", 251.0, "AWACS frequency in MHz")
	GCICallsign := flag.String("callsign", "", "GCI callsign. Used in radio transmissions")
	Coalition := flag.String("coalition", "blue", "Coalition (either blue or red)")
	RadarSweepInterval := flag.Duration("radar-sweep-interval", 2*time.Second, "Radar update tick rate")
	WhisperModelPath := flag.String("whisper-model", "", "Path to whisper.cpp model")
	Voice := flag.String("voice", "", "Voice to use for SRS transmissions (feminine, masculine)")
	ShiftLength := flag.Duration("shift-length", 8*time.Hour, "Bot will internally restart on this interval")

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

	hour := time.Now().Hour()
	seed := time.Now().YearDay()
	if 0 <= hour && hour < 8 {
		seed = seed + 1
	} else if 8 <= hour && hour < 16 {
		seed = seed * 2
	} else if 16 <= hour && hour < 24 {
		seed = seed - 3
	}
	rando := rand.New(rand.NewSource(int64(seed)))

	var voice voices.Voice
	switch strings.ToLower(*Voice) {
	case "":
		voices := []voices.Voice{voices.FeminineVoice, voices.MasculineVoice}
		voice = voices[rando.Intn(len(voices))]
		log.Info().Type("voice", voice).Msg("randomly selected voice")
	case "feminine":
		log.Info().Msg("using feminine voice")
		voice = voices.FeminineVoice
	case "masculine":
		log.Info().Msg("using masculine voice")
		voice = voices.MasculineVoice
	default:
		err = fmt.Errorf("unknown voice: %s", *Voice)
		exitOnErr(err)
	}

	frequency := unit.Frequency(*SRSFrequency) * unit.Megahertz

	callsign := *GCICallsign
	if callsign == "" {
		callsign = conf.DefaultCallsigns[rando.Intn(len(conf.DefaultCallsigns))]
		log.Info().Str("callsign", callsign).Msg("randomly selected callsign")
	}

	// Configure and run the application.
	config := conf.Configuration{
		ACMIFile:                     *ACMIFile,
		TelemetryAddress:             *TelemetryAddress,
		TelemetryConnectionTimeout:   *TelemetryConnectionTimeout,
		TelemetryClientName:          callsign,
		TelemetryPassword:            *TelemetryPassword,
		SRSAddress:                   *SRSAddress,
		SRSConnectionTimeout:         *SRSConnectionTimeout,
		SRSClientName:                fmt.Sprintf("GCI %s [BOT]", callsign),
		SRSExternalAWACSModePassword: *SRSExternalAWACSModePassword,
		SRSFrequency:                 frequency,
		Callsign:                     callsign,
		Coalition:                    coalition,
		RadarSweepInterval:           *RadarSweepInterval,
		WhisperModel:                 whisperModel,
		Voice:                        voice,
	}

	log.Info().Msg("starting application")
	for {
		runCtx, cancel := context.WithTimeout(ctx, *ShiftLength)
		err := runApplication(runCtx, cancel, config)
		exitOnErr(err)
		time.Sleep(5 * time.Second)
	}
}

func runApplication(ctx context.Context, cancel context.CancelFunc, config conf.Configuration) error {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Any("recovered", r).Msg("!!! APPLICATION PANIC RECOVERY !!!")
		}
	}()
	log.Info().Msg("starting new application instance")
	app, err := application.NewApplication(ctx, config)
	if err != nil {
		return err
	}
	err = app.Run(ctx, cancel)
	if err != nil {
		log.Error().Err(err).Msg("application exited with error")
	}
	return nil
}

// exitOnErr logs the error and exits the application if the error is not nil.
func exitOnErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("application exiting with error")
		os.Exit(1)
	}
}
