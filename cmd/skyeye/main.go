package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/dharmab/skyeye/internal/application"
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Variables for CLI/Config flags
var (
	configFile                   string
	logLevel                     string
	logFormat                    string
	acmiFile                     string
	telemetryAddress             string
	telemetryConnectionTimeout   time.Duration
	telemetryPassword            string
	srsAddress                   string
	srsConnectionTimeout         time.Duration
	srsExternalAWACSModePassword string
	srsFrequency                 float64
	gciCallsign                  string
	gciCallsigns                 []string
	coalitionName                string
	telemetryUpdateInterval      time.Duration
	whisperModelPath             string
	voiceName                    string
	playbackSpeed                string
	enableAutomaticPicture       bool
	automaticPictureInterval     time.Duration
	enableThreatMonitoring       bool
	threatMonitoringInterval     time.Duration
	threatMonitoringRequiresSRS  bool
	mandatoryThreatRadiusNM      float64
)

var (
	envPrefix = "SKYEYE"
)

func init() {
	skyeye.Flags().StringVar(&configFile, "config-file", ".", "Path to a config file e.g. '/home/user/xyz.yaml'")

	// Logging
	logLevelFlag := NewEnum(&logLevel, "Level", "info", "error", "warn", "info", "debug", "trace")
	skyeye.Flags().Var(logLevelFlag, "log-level", "Log level (error, warn, info, debug, trace)")
	logFormats := NewEnum(&logFormat, "Format", "pretty", "json")
	skyeye.Flags().Var(logFormats, "log-format", "Log format (pretty, json)")

	// Telemetry
	skyeye.Flags().StringVar(&acmiFile, "acmi-file", "", "path to ACMI file")
	skyeye.Flags().StringVar(&telemetryAddress, "telemetry-address", "localhost:42674", "Address of the real-time telemetry service")
	skyeye.MarkFlagsMutuallyExclusive("acmi-file", "telemetry-address")
	skyeye.Flags().DurationVar(&telemetryConnectionTimeout, "telemetry-connection-timeout", 10*time.Second, "Connection timeout for real-time telemetry client")
	skyeye.Flags().StringVar(&telemetryPassword, "telemetry-password", "", "Password for the real-time telemetry service")
	skyeye.Flags().DurationVar(&telemetryUpdateInterval, "telemetry-update-interval", 2*time.Second, "Interval at which trackfiles are updated from telemetry data")

	// SRS
	skyeye.Flags().StringVar(&srsAddress, "srs-server-address", "localhost:5002", "Address of the SRS server")
	skyeye.Flags().DurationVar(&srsConnectionTimeout, "srs-connection-timeout", 10*time.Second, "Connection timeout for SRS client")
	skyeye.Flags().StringVar(&srsExternalAWACSModePassword, "srs-eam-password", "", "SRS external AWACS mode password")
	skyeye.Flags().Float64Var(&srsFrequency, "srs-frequency", 251.0, "AWACS frequency in MHz")

	// Identity
	skyeye.Flags().StringVar(&gciCallsign, "callsign", "", "GCI callsign used in radio transmissions. Automatically chosen if not provided.")
	skyeye.Flags().StringSliceVar(&gciCallsigns, "callsigns", []string{}, "A list of GCI callsigns to select from.")
	skyeye.MarkFlagsMutuallyExclusive("callsign", "callsigns")
	coalitionFlag := NewEnum(&coalitionName, "Coalition", "blue", "red")
	skyeye.Flags().Var(coalitionFlag, "coalition", "GCI coalition (blue, red)")

	// AI models
	skyeye.Flags().StringVar(&whisperModelPath, "whisper-model", "", "Path to whisper.cpp model")
	_ = skyeye.MarkFlagRequired("whisper-model")
	voiceFlag := NewEnum(&voiceName, "Voice", "", "feminine", "masculine")
	skyeye.Flags().Var(voiceFlag, "voice", "Voice to use for SRS transmissions (feminine, masculine)")
	playbackSpeedFlag := NewEnum(&playbackSpeed, "string", "standard", "veryslow", "slow", "fast", "veryfast")
	skyeye.Flags().Var(playbackSpeedFlag, "voice-playback-speed", "Voice playback speed of GCI")

	// Controller behavior
	skyeye.Flags().BoolVar(&enableAutomaticPicture, "auto-picture", true, "Enable automatic PICTURE broadcasts")
	skyeye.Flags().DurationVar(&automaticPictureInterval, "auto-picture-interval", 2*time.Minute, "How often to broadcast PICTURE")
	skyeye.Flags().BoolVar(&enableThreatMonitoring, "threat-monitoring", true, "Enable THREAT monitoring")
	skyeye.Flags().DurationVar(&threatMonitoringInterval, "threat-monitoring-interval", 3*time.Minute, "How often to broadcast THREAT")
	skyeye.Flags().Float64Var(&mandatoryThreatRadiusNM, "mandatory-threat-radius", 25, "Briefed radius for mandatory THREAT calls, in nautical miles")
	skyeye.Flags().BoolVar(&threatMonitoringRequiresSRS, "threat-monitoring-requires-srs", true, "Require aircraft to be on SRS to receive THREAT calls. Only useful to disable when debugging.")
}

// Top-level CLI command
var skyeye = &cobra.Command{
	Use:     "skyeye",
	Version: Version,
	Short:   "AI Powered GCI Bot for DCS World",
	Long:    "Skyeye uses real-time telemetry data from TacView to provide Ground-Controlled Intercept service over SimpleRadio-Standalone.",
	Example: strings.Join(
		[]string{
			"  " + "Custom Config Path",
			"skyeye --config-file='/home/user/xyz.yaml'",
			"",
			"Remote TacView and SRS server",
			"skyeye --telemetry-address=your-tacview-server:42674 --telemetry-password=your-tacview-password --srs-server-address=your-srs-server:5002 --srs-eam-password=your-srs-eam-password --whisper-model=ggml-small.en.bin",
			"",
			"Local TacView and SRS server",
			"skyeye --telemetry-password=your-tacview-password --srs-eam-password=your-srs-eam-password --whisper-model=ggml-small.en.bin",
		},
		"\n  ",
	),
	PreRun: func(cmd *cobra.Command, args []string) {
		initializeConfig(cmd)

		if whisperModelPath == "" && !viper.IsSet("whisper-model") {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
	Run: Supervise,
}

func main() {
	cobra.MousetrapHelpText = "Thanks for trying SkyEye! SkyEye is a command-line application. Please read the documentation on GitHub for instructions on how to run it, or run the program from a terminal to see more help text. "
	cobra.MousetrapDisplayDuration = 0
	if err := skyeye.Execute(); err != nil {
		log.Error().Err(err).Msg("application exited with error")
		os.Exit(1)
	}
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		// having no config file is fine
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	bindFlags(cmd, v)
	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := f.Name

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

func setupLogging() {
	if strings.EqualFold(logFormat, "pretty") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	var level zerolog.Level
	switch strings.ToLower(logLevel) {
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
	log.Info().Stringer("level", level).Msg("log level set")
}

func loadCoalition() (coalition coalitions.Coalition) {
	log.Info().Str("coalition", coalitionName).Msg("setting GCI coalition")
	switch coalitionName {
	case "blue":
		coalition = coalitions.Blue
	case "red":
		coalition = coalitions.Red
	default:
		exitOnErr(errors.New("GCI coalition must be either blue or red"))
	}
	log.Info().Int("id", int(coalition)).Msg("GCI coalition set")
	return
}

func loadPlaybackSpeed() float32 {
	speedMap := map[string]float32{
		"veryslow": 1.3,
		"slow":     1.15,
		"standard": 1.0,
		"fast":     0.85,
		"veryfast": 0.7,
	}
	if speed, ok := speedMap[playbackSpeed]; ok {
		log.Info().Float32("speed", speed).Msg("setting playback speed")
		return speed
	} else {
		log.Info().Float32("speed", speed).Msg("Unknown playback speed, revert to default (standard)")
		return 1.0
	}
}

func loadWhisperModel() *whisper.Model {
	log.Info().Str("path", whisperModelPath).Msg("loading whisper model")
	whisperModel, err := whisper.New(whisperModelPath)
	if err != nil {
		exitOnErr(fmt.Errorf("failed to load whisper model: %w", err))
	}
	log.Info().Msg("whisper model loaded")
	return &whisperModel
}

func randomizer() (rando *rand.Rand) {
	hour := time.Now().Hour()
	seed := time.Now().YearDay()
	if 0 <= hour && hour < 8 {
		seed = seed + 1
	} else if 8 <= hour && hour < 16 {
		seed = seed * 2
	} else if 16 <= hour && hour < 24 {
		seed = seed - 3
	}
	rando = rand.New(rand.NewSource(int64(seed)))
	return
}

func loadVoice(rando *rand.Rand) (voice voices.Voice) {
	options := map[string]voices.Voice{
		"feminine":  voices.FeminineVoice,
		"masculine": voices.MasculineVoice,
	}
	if voiceName == "" {
		keys := reflect.ValueOf(options).MapKeys()
		voice = options[keys[rando.Intn(len(keys))].String()]
		log.Info().Type("voice", voice).Msg("randomly selected voice")
	} else {
		voice = options[voiceName]
		log.Info().Type("voice", voice).Msg("selected voice")
	}
	return
}

func loadCallsign(rando *rand.Rand) (callsign string) {
	var options []string
	if gciCallsign != "" {
		options = append(options, gciCallsign)
	}
	if len(gciCallsigns) > 0 {
		options = append(options, gciCallsigns...)
	}
	if len(options) == 0 {
		options = conf.DefaultCallsigns
	}
	callsign = options[rando.Intn(len(options))]
	if callsign == "" {
		panic("callsign is empty")
	}
	log.Info().Str("callsign", callsign).Msg("selected callsign")
	return
}

func loadFrequency() (frequency unit.Frequency) {
	frequency = unit.Frequency(srsFrequency) * unit.Megahertz
	log.Info().Float64("frequency", frequency.Megahertz()).Msg("parsed SRS frequency")
	return
}

func Supervise(cmd *cobra.Command, args []string) {
	// Set up an application-scoped context and a cancel function to shut down the application.
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Set up logging
	setupLogging()

	log.Info().Msg("setting up interrupt and TERM signal handler")
	interuptChan := make(chan os.Signal, 1)
	signal.Notify(interuptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-interuptChan
		log.Info().Any("signal", s).Msg("received shutdown signal")
		cancel()
		go func() {
			time.Sleep(10 * time.Second)
			log.Warn().Msg("shutdown took too long, forcing exit")
			os.Exit(1)
		}()
		wg.Wait()
		os.Exit(0)
	}()

	log.Info().Msg("loading configuration")
	coalition := loadCoalition()
	whisperModel := loadWhisperModel()
	rando := randomizer()
	voice := loadVoice(rando)
	callsign := loadCallsign(rando)
	frequency := loadFrequency()
	playbackSpeed := loadPlaybackSpeed()

	config := conf.Configuration{
		ACMIFile:                     acmiFile,
		TelemetryAddress:             telemetryAddress,
		TelemetryConnectionTimeout:   telemetryConnectionTimeout,
		TelemetryClientName:          callsign,
		TelemetryPassword:            telemetryPassword,
		SRSAddress:                   srsAddress,
		SRSConnectionTimeout:         srsConnectionTimeout,
		SRSClientName:                fmt.Sprintf("GCI %s [BOT]", callsign),
		SRSExternalAWACSModePassword: srsExternalAWACSModePassword,
		SRSFrequency:                 frequency,
		Callsign:                     callsign,
		Coalition:                    coalition,
		RadarSweepInterval:           telemetryUpdateInterval,
		WhisperModel:                 whisperModel,
		Voice:                        voice,
		PlaybackSpeed:                playbackSpeed,
		EnableThreatMonitoring:       enableThreatMonitoring,
		ThreatMonitoringInterval:     threatMonitoringInterval,
		ThreatMonitoringRequiresSRS:  threatMonitoringRequiresSRS,
		MandatoryThreatRadius:        unit.Length(mandatoryThreatRadiusNM) * unit.NauticalMile,
	}

	if enableAutomaticPicture {
		config.PictureBroadcastInterval = automaticPictureInterval
		log.Info().Dur("interval", automaticPictureInterval).Msg("automatic PICTURE broadcasts enabled")
	} else {
		config.PictureBroadcastInterval = 117 * time.Hour
	}

	log.Info().Msg("starting application")
	err := runApplication(ctx, &wg, config)
	exitOnErr(err)
}

func runApplication(ctx context.Context, wg *sync.WaitGroup, config conf.Configuration) error {
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
	err = app.Run(ctx, wg)
	if err != nil {
		log.Error().Err(err).Msg("application exited with error")
	}
	wg.Wait()
	return nil
}

// exitOnErr logs the error and exits the application if the error is not nil.
func exitOnErr(err error) {
	if err != nil {
		log.Error().Err(err).Msg("application exiting with error")
		os.Exit(1)
	}
}
