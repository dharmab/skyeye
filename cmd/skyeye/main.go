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
	"github.com/spf13/viper"

	"github.com/dharmab/skyeye/internal/application"
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Variables for CLI/Config flags
var (
	configPath                   string
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

func init() {
	skyeye.Flags().StringVar(&configPath, "config-path", ".", "Path to a config file e.g. '/home/user/xyz'. It is looking for a file called skyeye-config.yaml")

	viper.SetConfigName("skyeye-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Error().Err(err).Msg("Unable to read config yaml file")
	}

	// Logging
	logLevelFlag := NewEnum(&logLevel, "Level", "info", "error", "warn", "info", "debug", "trace")
	skyeye.Flags().Var(logLevelFlag, "log-level", "Log level (error, warn, info, debug, trace)")
	viper.BindPFlag("log-level", skyeye.Flags().Lookup("log-level"))
	viper.SetDefault("log-level", "info")

	logFormats := NewEnum(&logFormat, "Format", "pretty", "json")
	skyeye.Flags().Var(logFormats, "log-format", "Log format (pretty, json)")
	viper.BindPFlag("log-format", skyeye.Flags().Lookup("log-format"))
	viper.SetDefault("log-format", "pretty")

	// Telemetry
	skyeye.Flags().StringVar(&acmiFile, "acmi-file", "", "path to ACMI file")
	viper.BindPFlag("acmi-file", skyeye.Flags().Lookup("acmi-file"))

	skyeye.Flags().StringVar(&telemetryAddress, "telemetry-address", "localhost:42674", "Address of the real-time telemetry service")
	viper.BindPFlag("telemetry-address", skyeye.Flags().Lookup("telemetry-address"))
	viper.SetDefault("telemetry-address", "localhost:42674")

	skyeye.MarkFlagsMutuallyExclusive("acmi-file", "telemetry-address")

	skyeye.Flags().DurationVar(&telemetryConnectionTimeout, "telemetry-connection-timeout", 10*time.Second, "Connection timeout for real-time telemetry client")
	viper.BindPFlag("telemetry-connection-timeout", skyeye.Flags().Lookup("telemetry-connection-timeout"))
	viper.SetDefault("telemetry-connection-timeout", 10*time.Second)

	skyeye.Flags().StringVar(&telemetryPassword, "telemetry-password", "", "Password for the real-time telemetry service")
	viper.BindPFlag("telemetry-password", skyeye.Flags().Lookup("telemetry-password"))

	skyeye.Flags().DurationVar(&telemetryUpdateInterval, "telemetry-update-interval", 2*time.Second, "Interval at which trackfiles are updated from telemetry data")
	viper.BindPFlag("telemetry-update-interval", skyeye.Flags().Lookup("telemetry-update-interval"))
	viper.SetDefault("telemetry-update-interval", 2*time.Second)

	// SRS
	skyeye.Flags().StringVar(&srsAddress, "srs-server-address", "localhost:5002", "Address of the SRS server")
	viper.BindPFlag("srs-server-address", skyeye.Flags().Lookup("srs-server-address"))
	viper.SetDefault("srs-server-address", "localhost:5002")

	skyeye.Flags().DurationVar(&srsConnectionTimeout, "srs-connection-timeout", 10*time.Second, "Connection timeout for SRS client")
	viper.BindPFlag("srs-connection-timeout", skyeye.Flags().Lookup("srs-connection-timeout"))
	viper.SetDefault("srs-connection-timeout", 10*time.Second)

	skyeye.Flags().StringVar(&srsExternalAWACSModePassword, "srs-eam-password", "", "SRS external AWACS mode password")
	viper.BindPFlag("srs-eam-password", skyeye.Flags().Lookup("srs-eam-password"))

	skyeye.Flags().Float64Var(&srsFrequency, "srs-frequency", 251.0, "AWACS frequency in MHz")
	viper.BindPFlag("srs-frequency", skyeye.Flags().Lookup("srs-frequency"))
	viper.SetDefault("srs-frequency", 251.0)

	// Identity
	skyeye.Flags().StringVar(&gciCallsign, "callsign", "", "GCI callsign used in radio transmissions. Automatically chosen if not provided.")
	viper.BindPFlag("callsign", skyeye.Flags().Lookup("callsign"))

	skyeye.Flags().StringSliceVar(&gciCallsigns, "callsigns", []string{}, "A list of GCI callsigns to select from.")
	viper.BindPFlag("callsigns", skyeye.Flags().Lookup("callsigns"))

	skyeye.MarkFlagsMutuallyExclusive("callsign", "callsigns")

	coalitionFlag := NewEnum(&coalitionName, "Coalition", "blue", "red")
	skyeye.Flags().Var(coalitionFlag, "coalition", "GCI coalition (blue, red)")
	viper.BindPFlag("coalition", skyeye.Flags().Lookup("coalition"))

	// AI models
	skyeye.Flags().StringVar(&whisperModelPath, "whisper-model", "", "Path to whisper.cpp model")
	viper.BindPFlag("whisper-model", skyeye.Flags().Lookup("whisper-model"))

	voiceFlag := NewEnum(&voiceName, "Voice", "", "feminine", "masculine")
	skyeye.Flags().Var(voiceFlag, "voice", "Voice to use for SRS transmissions (feminine, masculine)")
	viper.BindPFlag("voice", skyeye.Flags().Lookup("voice"))

	playbackSpeedFlag := NewEnum(&playbackSpeed, "string", "standard", "veryslow", "slow", "fast", "veryfast")
	skyeye.Flags().Var(playbackSpeedFlag, "voice-playback-speed", "Voice playback speed of GCI")
	viper.BindPFlag("voice-playback-speed", skyeye.Flags().Lookup("voice-playback-speed"))

	// Controller behavior
	skyeye.Flags().BoolVar(&enableAutomaticPicture, "auto-picture", true, "Enable automatic PICTURE broadcasts")
	viper.BindPFlag("auto-picture", skyeye.Flags().Lookup("auto-picture"))
	viper.SetDefault("auto-picture", true)

	skyeye.Flags().DurationVar(&automaticPictureInterval, "auto-picture-interval", 2*time.Minute, "How often to broadcast PICTURE")
	viper.BindPFlag("auto-picture-interval", skyeye.Flags().Lookup("auto-picture-interval"))
	viper.SetDefault("auto-picture-interval", 2*time.Minute)

	skyeye.Flags().BoolVar(&enableThreatMonitoring, "threat-monitoring", true, "Enable THREAT monitoring")
	viper.BindPFlag("threat-monitoring", skyeye.Flags().Lookup("threat-monitoring"))
	viper.SetDefault("threat-monitoring", true)

	skyeye.Flags().DurationVar(&threatMonitoringInterval, "threat-monitoring-interval", 3*time.Minute, "How often to broadcast THREAT")
	viper.BindPFlag("threat-monitoring-interval", skyeye.Flags().Lookup("threat-monitoring-interval"))
	viper.SetDefault("threat-monitoring-interval", 3*time.Minute)

	skyeye.Flags().Float64Var(&mandatoryThreatRadiusNM, "mandatory-threat-radius", 25, "Briefed radius for mandatory THREAT calls, in nautical miles")
	viper.BindPFlag("mandatory-threat-radius", skyeye.Flags().Lookup("mandatory-threat-radius"))
	viper.SetDefault("mandatory-threat-radius", 25.0)

	skyeye.Flags().BoolVar(&threatMonitoringRequiresSRS, "threat-monitoring-requires-srs", true, "Require aircraft to be on SRS to receive THREAT calls. Only useful to disable when debugging.")
	viper.BindPFlag("threat-monitoring-requires-srs", skyeye.Flags().Lookup("threat-monitoring-requires-srs"))
	viper.SetDefault("threat-monitoring-requires-srs", true)
}

// Top-level CLI command
var skyeye = &cobra.Command{
	Use:     "skyeye",
	Version: Version,
	Short:   "AI Powered GCI Bot for DCS World",
	Long:    "Skyeye uses real-time telemetry data from TacView to provide Ground-Controlled Intercept service over SimpleRadio-Standalone.",
	Example: strings.Join(
		[]string{
			"Custom Config Path",
			"skyeye --config-path='/home/user/xyz'",
			"",
			"  " + "Remote TacView and SRS server",
			"skyeye --telemetry-address=your-tacview-server:42674 --telemetry-password=your-tacview-password --srs-server-address=your-srs-server:5002 --srs-eam-password=your-srs-eam-password --whisper-model=ggml-small.en.bin",
			"",
			"Local TacView and SRS server",
			"skyeye --telemetry-password=your-tacview-password --srs-eam-password=your-srs-eam-password --whisper-model=ggml-small.en.bin",
		},
		"\n  ",
	),
	PreRun: func(cmd *cobra.Command, args []string) {
		if whisperModelPath == "" && !viper.IsSet("whisper-model") {
			_ = cmd.Help()
			os.Exit(0)
		}
		// Load all necessary config parameters from Viper
		logLevel = viper.GetString("log-level")
		logFormat = viper.GetString("log-format")
		acmiFile = viper.GetString("acmi-file")
		telemetryAddress = viper.GetString("telemetry-address")
		telemetryConnectionTimeout = viper.GetDuration("telemetry-connection-timeout")
		telemetryPassword = viper.GetString("telemetry-password")
		telemetryUpdateInterval = viper.GetDuration("telemetry-update-interval")
		srsAddress = viper.GetString("srs-server-address")
		srsConnectionTimeout = viper.GetDuration("srs-connection-timeout")
		srsExternalAWACSModePassword = viper.GetString("srs-eam-password")
		srsFrequency = viper.GetFloat64("srs-frequency")
		gciCallsign = viper.GetString("callsign")
		gciCallsigns = viper.GetStringSlice("callsigns")
		coalitionName = viper.GetString("coalition")
		whisperModelPath = viper.GetString("whisper-model")
		voiceName = viper.GetString("voice")
		playbackSpeed = viper.GetString("voice-playback-speed")
		enableAutomaticPicture = viper.GetBool("auto-picture")
		automaticPictureInterval = viper.GetDuration("auto-picture-interval")
		enableThreatMonitoring = viper.GetBool("threat-monitoring")
		threatMonitoringInterval = viper.GetDuration("threat-monitoring-interval")
		mandatoryThreatRadiusNM = viper.GetFloat64("mandatory-threat-radius")
		threatMonitoringRequiresSRS = viper.GetBool("threat-monitoring-requires-srs")
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
