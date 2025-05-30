package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/cpu"

	"github.com/gofrs/flock"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/dharmab/skyeye/internal/application"
	"github.com/dharmab/skyeye/internal/cli"
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Used for CLI configuration values.
var (
	configFile                   string
	logLevel                     string
	logFormat                    string
	enableTranscriptionLogging   bool
	acmiFile                     string
	telemetryAddress             string
	telemetryConnectionTimeout   time.Duration
	telemetryPassword            string
	srsAddress                   string
	srsConnectionTimeout         time.Duration
	srsExternalAWACSModePassword string
	srsFrequencies               []string
	enableGRPC                   bool
	grpcAddress                  string
	grpcAPIKey                   string
	gciCallsign                  string
	gciCallsigns                 []string
	coalitionName                string
	telemetryUpdateInterval      time.Duration
	recognizerName               string
	whisperModelPath             string
	recognizerLockPath           string
	openAIAPIKey                 string
	voiceName                    string
	useSystemVoice               bool
	mute                         bool
	voiceSpeed                   float64
	voicePauseLength             time.Duration
	voiceLockPath                string
	enableAutomaticPicture       bool
	automaticPictureInterval     time.Duration
	enableThreatMonitoring       bool
	threatMonitoringInterval     time.Duration
	threatMonitoringRequiresSRS  bool
	mandatoryThreatRadiusNM      float64
	enableTracing                bool
	discordWebhookID             string
	discordWebhookToken          string
	exitAfter                    time.Duration
)

func init() {
	skyeye.Flags().StringVar(&configFile, "config-file", "/etc/skyeye/config.yaml", "Path to config file")

	// Logging
	logLevelFlag := cli.NewEnum(&logLevel, "Level", "info", "error", "warn", "info", "debug", "trace")
	skyeye.Flags().Var(logLevelFlag, "log-level", "Log level (error, warn, info, debug, trace)")
	logFormats := cli.NewEnum(&logFormat, "Format", "pretty", "json")
	skyeye.Flags().Var(logFormats, "log-format", "Log format (pretty, json)")
	skyeye.Flags().BoolVar(&enableTranscriptionLogging, "enable-transcription-logging", true, "Include transcriptions of SRS transmissions in logs and traces")

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
	skyeye.Flags().StringSliceVar(&srsFrequencies, "srs-frequencies", []string{"251.0AM", "133.0AM", "30.0FM"}, "List of SRS frequencies to use")

	// DCS-gRPC
	skyeye.Flags().BoolVar(&enableGRPC, "enable-grpc", false, "Enable DCS-gRPC features")
	skyeye.Flags().StringVar(&grpcAddress, "grpc-address", "localhost:50051", "Address of the DCS-gRPC server")
	skyeye.Flags().StringVar(&grpcAPIKey, "grpc-api-key", "", "API key for DCS-gRPC authentication")

	// Identity
	skyeye.Flags().StringVar(&gciCallsign, "callsign", "", "GCI callsign used in radio transmissions. Automatically chosen if not provided")
	skyeye.Flags().StringSliceVar(&gciCallsigns, "callsigns", []string{}, "A list of GCI callsigns to select from")
	skyeye.MarkFlagsMutuallyExclusive("callsign", "callsigns")
	coalitionFlag := cli.NewEnum(&coalitionName, "Coalition", "blue", "red")
	skyeye.Flags().Var(coalitionFlag, "coalition", "GCI coalition (blue, red)")

	// Speech-to-text
	recognizerFlag := cli.NewEnum(&recognizerName, "Recognizer", string(conf.WhisperLocal), string(conf.WhisperAPI), string(conf.GPT4o), string(conf.GPT4oMini))
	skyeye.Flags().Var(recognizerFlag, "recognizer", "Speech-to-text recognizer to use")
	skyeye.Flags().StringVar(&whisperModelPath, "whisper-model", "", "Path to whisper.cpp model")
	skyeye.Flags().StringVar(&openAIAPIKey, "openai-api-key", "", "API key for OpenAPI AI")
	skyeye.MarkFlagsOneRequired("whisper-model", "openai-api-key")
	skyeye.Flags().StringVar(&recognizerLockPath, "recognizer-lock-path", "", "Path to lock file for concurrent speech-to-text when using multiple instances")

	// Text-to-speech
	voiceFlag := cli.NewEnum(&voiceName, "Voice", "", "feminine", "masculine")
	skyeye.Flags().Var(voiceFlag, "voice", "Voice to use for SRS transmissions (feminine, masculine). Automatically chosen if not provided.")
	skyeye.Flags().Float64Var(&voiceSpeed, "voice-playback-speed", 1.0, "How quickly the GCI speaks (values below 1.0 are faster and above are slower).")
	skyeye.Flags().BoolVar(&mute, "mute", false, "Mute all SRS transmissions. Useful for testing without disrupting play")
	skyeye.Flags().StringVar(&voiceLockPath, "voice-lock-path", "", "Path to lock file for concurrent text-to-speech when using multiple instances")
	if runtime.GOOS == "darwin" {
		skyeye.Flags().BoolVar(&useSystemVoice, "use-system-voice", false, "Use the System Voice chosen in the Spoken Content page in System Settings instead of Samantha.")
		if err := skyeye.Flags().MarkDeprecated("voice", "Select a voice in System Settings and use --use-system-voice instead."); err != nil {
			log.Fatal().Err(err).Msg("failed to mark flag as deprecated")
		}
	} else {
		skyeye.Flags().DurationVar(&voicePauseLength, "voice-playback-pause", 200*time.Millisecond, "How long the GCI pauses between sentences.")
	}

	// Controller behavior
	skyeye.Flags().BoolVar(&enableAutomaticPicture, "auto-picture", true, "Enable automatic PICTURE broadcasts")
	skyeye.Flags().DurationVar(&automaticPictureInterval, "auto-picture-interval", 2*time.Minute, "How often to broadcast PICTURE")
	skyeye.Flags().BoolVar(&enableThreatMonitoring, "threat-monitoring", true, "Enable THREAT monitoring")
	skyeye.Flags().DurationVar(&threatMonitoringInterval, "threat-monitoring-interval", 3*time.Minute, "How often to broadcast THREAT")
	skyeye.Flags().Float64Var(&mandatoryThreatRadiusNM, "mandatory-threat-radius", 25, "Briefed radius for mandatory THREAT calls, in nautical miles")
	skyeye.Flags().BoolVar(&threatMonitoringRequiresSRS, "threat-monitoring-requires-srs", true, "Require aircraft to be on SRS to receive THREAT calls. Only useful to disable when debugging")

	// Tracing
	skyeye.Flags().BoolVar(&enableTracing, "enable-tracing", false, "Enable tracing")
	skyeye.Flags().BoolVar(&enableTracing, "tracing", false, "Enable tracing")
	if err := skyeye.Flags().MarkDeprecated("tracing", "use --enable-tracing instead"); err != nil {
		log.Fatal().Err(err).Msg("failed to mark flag as deprecated")
	}
	skyeye.MarkFlagsMutuallyExclusive("tracing", "enable-tracing")
	skyeye.Flags().StringVar(&discordWebhookID, "discord-webhook-id", "", "Discord webhook ID for tracing")
	skyeye.Flags().StringVar(&discordWebhookToken, "discord-webhook-token", "", "Discord webhook token for tracing")
	skyeye.MarkFlagsRequiredTogether("discord-webhook-id", "discord-webhook-token")

	// Runtime
	skyeye.Flags().DurationVar(&exitAfter, "exit-after", time.Hour*24*365*20, "Exit after running for the specified duration")
}

// Top-level CLI command.
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
	PreRunE: preRun,
	Run:     run,
}

func main() {
	cobra.MousetrapHelpText = "Thanks for trying SkyEye! SkyEye is a command-line application. Please read the documentation on GitHub for instructions on how to run it, or run the program from a terminal to see more help text. "
	cobra.MousetrapDisplayDuration = 0
	if err := skyeye.Execute(); err != nil {
		log.Fatal().Err(err).Msg("application exited with error")
	}
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigFile(configFile)
	if err := v.ReadInConfig(); err != nil {
		// having no config file is fine
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	v.SetEnvPrefix("SKYEYE")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	bindFlags(cmd, v)
	return nil
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprint(val)); err != nil {
				log.Warn().Str("flag", f.Name).Msg("Failed to set flag")
			}
		}
	})
}

func loadCoalition() (coalition coalitions.Coalition) {
	log.Info().Str("coalition", coalitionName).Msg("setting GCI coalition")
	switch coalitionName {
	case "blue":
		coalition = coalitions.Blue
	case "red":
		coalition = coalitions.Red
	default:
		log.Fatal().Msg("GCI coalition must be either blue or red")
	}
	log.Info().Int("id", int(coalition)).Msg("GCI coalition set")
	return
}

func loadWhisperModel() *whisper.Model {
	if recognizerName != string(conf.WhisperLocal) {
		return nil
	}
	if whisperModelPath == "" {
		log.Fatal().Msg("whisper-model is required when recognizer is set to " + string(conf.WhisperLocal))
	}
	if runtime.GOARCH == "amd64" && !cpu.X86.HasAVX2 {
		log.Fatal().Msg("The CPU on this machine does not support AVX2 instructions.")
	}

	log.Info().Str("path", whisperModelPath).Msg("loading whisper model")
	whisperModel, err := whisper.New(whisperModelPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", whisperModelPath).Err(err).Msg("failed to load whisper model")
	}
	log.Info().
		Bool("multilingual", whisperModel.IsMultilingual()).
		Strs("languages", whisperModel.Languages()).
		Msg("whisper model loaded")
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
	rando = rand.New(rand.NewPCG(uint64(seed), uint64(seed))) // nolint:gosec // seed is bounded to small values, so overflow is impossible
	return
}

func loadVoice(rando *rand.Rand) (voice voices.Voice) {
	options := map[string]voices.Voice{
		"feminine":  voices.FeminineVoice,
		"masculine": voices.MasculineVoice,
	}
	if voiceName == "" {
		keys := reflect.ValueOf(options).MapKeys()
		voice = options[keys[rando.IntN(len(keys))].String()]
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
	callsign = options[rando.IntN(len(options))]
	if callsign == "" {
		panic("callsign is empty")
	}
	log.Info().Str("callsign", callsign).Msg("selected callsign")
	return
}

func loadLock(path string) *flock.Flock {
	if path == "" {
		return nil
	}
	log.Info().Str("path", path).Msg("using lock file")
	return flock.New(path)
}

func preRun(cmd *cobra.Command, _ []string) error {
	if err := initializeConfig(cmd); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}
	return nil
}

func run(_ *cobra.Command, _ []string) {
	// Set up an application-scoped context and a cancel function to shut down the application.
	ctx, cancel := context.WithCancel(context.Background())

	// Safety in case of hung routine
	go func() {
		<-ctx.Done()
		time.Sleep(10 * time.Second)
		log.Warn().Msg("shutdown took too long, forcing exit")
		_ = pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		os.Exit(1)
	}()

	var wg sync.WaitGroup

	cli.SetupZerolog(logLevel, logFormat)

	log.Info().Str("version", Version).Msg("SkyEye GCI Bot")

	log.Info().Msg("setting up interrupt and TERM signal handler")
	interuptChan := make(chan os.Signal, 1)
	signal.Notify(interuptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-interuptChan
		log.Info().Any("signal", s).Msg("received shutdown signal")
		cancel()
		wg.Wait()
		os.Exit(0)
	}()

	log.Info().Msg("loading configuration")
	coalition := loadCoalition()
	whisperModel := loadWhisperModel()
	rando := randomizer()
	voice := loadVoice(rando)
	callsign := loadCallsign(rando)
	parsedSRSFrequencies := cli.LoadFrequencies(srsFrequencies)
	voiceLock := loadLock(voiceLockPath)
	recognizerLock := loadLock(recognizerLockPath)

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
		SRSFrequencies:               parsedSRSFrequencies,
		EnableTranscriptionLogging:   enableTranscriptionLogging,
		Callsign:                     callsign,
		Coalition:                    coalition,
		RadarSweepInterval:           telemetryUpdateInterval,
		Recognizer:                   conf.Recognizer(recognizerName),
		RecognizerLock:               recognizerLock,
		WhisperModel:                 whisperModel,
		OpenAIAPIKey:                 openAIAPIKey,
		Voice:                        voice,
		VoiceLock:                    voiceLock,
		Mute:                         mute,
		VoiceSpeed:                   voiceSpeed,
		VoicePauseLength:             voicePauseLength,
		EnableAutomaticPicture:       enableAutomaticPicture,
		PictureBroadcastInterval:     automaticPictureInterval,
		EnableThreatMonitoring:       enableThreatMonitoring,
		ThreatMonitoringInterval:     threatMonitoringInterval,
		ThreatMonitoringRequiresSRS:  threatMonitoringRequiresSRS,
		MandatoryThreatRadius:        unit.Length(mandatoryThreatRadiusNM) * unit.NauticalMile,
		EnableTracing:                enableTracing,
		DiscordWebhookID:             discordWebhookID,
		DiscorbWebhookToken:          discordWebhookToken,
		ExitAfter:                    exitAfter,
		EnableGRPC:                   enableGRPC,
		GRPCAddress:                  grpcAddress,
		GRPCAPIKey:                   grpcAPIKey,
	}

	log.Info().Msg("starting application")
	app, err := application.NewApplication(config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start application")
	}
	err = app.Run(ctx, cancel, &wg)
	if err != nil {
		log.Fatal().Err(err).Msg("application exited with error")
	}
	wg.Wait()
}
