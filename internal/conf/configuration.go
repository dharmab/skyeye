package conf

import (
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/gofrs/flock"
	"github.com/martinlindhe/unit"
)

type Recognizer string

const (
	WhisperLocal Recognizer = "openai-whisper-local"
	WhisperAPI   Recognizer = "openai-whisper-api"
	GPT4o        Recognizer = "openai-gpt4o"
	GPT4oMini    Recognizer = "openai-gpt4o-mini"
)

// Configuration for the SkyEye application.
type Configuration struct {
	// ACMIFile is the path to the ACMI file
	ACMIFile string
	// TelemetryAddress is the network address of the real-time telemetry server (including port)
	TelemetryAddress string
	// TelemetryConnectionTimeout is the connection timeout for connecting to the real-time telemetry server
	TelemetryConnectionTimeout time.Duration
	// TelemetryClientName is the client hostname used when handshaking with the real-time telemetry server
	TelemetryClientName string
	// TelemetryPassword is the password for connecting to the real-time telemetry server
	TelemetryPassword string
	// SRSAddress is the network address of the SimpleRadio Standalone server (including port)
	SRSAddress string
	// SRSConnectionTimeout is the connection timeout for connecting to the SimpleRadio Standalone server
	SRSConnectionTimeout time.Duration
	// SRSClientName is the name of the bot that will appear in the client list and in in-game transmissions
	SRSClientName string
	// SRSExternalAWACSModePassword is the password for connecting to the SimpleRadio Standalone server using External AWACS Mode
	SRSExternalAWACSModePassword string
	// SRSFrequencies that the bot simultaneously receives and transmits on
	SRSFrequencies []simpleradio.RadioFrequency
	// EnableGRPC controls whether DCS-gRPC features are enabled
	EnableGRPC bool
	// GRPCAddress is the network address of the DCS-gRPC server (including port)
	GRPCAddress string
	// GRPCAPIKey is the API key for authenticating with the DCS-gRPC server
	GRPCAPIKey string
	// EnableTranscriptionLogging controls whether transcriptions are included in logs.
	EnableTranscriptionLogging bool
	// Callsign is the GCI callsign used on SRS
	Callsign string
	// Coalition is the coalition that the bot will act on
	Coalition coalitions.Coalition
	// RadarSweepInterval is the rate at which the radar will update. This does not impact performance - ACMI data is still streamed at the same rate.
	// It only impacts the update rate of the GCI radar picture.
	RadarSweepInterval time.Duration
	// Recognizer selects which speech-to-text recognizer to use.
	Recognizer Recognizer
	// RecognizerLock is a file-based lock to control multiple instances running the recognizer at the same time.
	RecognizerLock *flock.Flock
	// WhisperModel is a whisper.cpp model used for Speech To Text. It may be nil if OpenAI API transcription is configured.
	WhisperModel *whisper.Model
	// OpenAIAPIKey is the API key for the OpenAI API. It may be empty if local transcription is configured.
	OpenAIAPIKey string
	// Voice is the voice used for SRS transmissions
	Voice voices.Voice
	// UseSystemVoice controls whether to use the System Voice on macOS. This allows use of current Siri voices,
	// but requires additional configuration in System Settings.
	UseSystemVoice bool
	// VoiceLock is a file-based lock to control multiple instances running Piper at the same time.
	VoiceLock *flock.Flock
	// Mute disables SRS transmissions
	Mute bool
	// Piper playback speed (default is 1.0) - The higher the value the slower it is.
	VoiceSpeed float64
	// Piper playback pause after every sentence in seconds (default is 0.2)
	VoicePauseLength time.Duration
	// EnableAutomaticPicture controls whether the controller will automatically broadcast a PICTURE at regular intervals.
	EnableAutomaticPicture bool
	// PictureBroadcastInterval is the interval at which the controller will automatically broadcast a PICTURE.
	PictureBroadcastInterval time.Duration
	// EnableThreatMonitoring controls whether the controller will broadcast THREAT calls.
	EnableThreatMonitoring bool
	// ThreatMonitoringInterval is the cooldown period between THREAT calls.
	ThreatMonitoringInterval time.Duration
	// MandatoryThreatRadius is the brief range at which a THREAT call is mandatory.
	MandatoryThreatRadius unit.Length
	// ThreatMonitoringRequiresSRS controls whether threat calls are issued to aircraft that are not on an SRS frequency. This is mostly
	// for debugging.
	ThreatMonitoringRequiresSRS bool
	// Locations is a slice of named locations that can be referenced in ALPHA CHECK and VECTOR calls.
	Locations []*Location
	// EnableTracing controls whether to publish traces
	EnableTracing bool
	// DiscordWebhookID is the ID of the Discord webhook
	DiscordWebhookID string
	// DiscordWebhookToken is the token for the Discord webhook
	DiscorbWebhookToken string
	// ExitAfter is the duration after which the application will exit
	ExitAfter time.Duration
}

type Location struct {
	// Names of the location
	Names []string `json:"names"`
	// Coordinates of the location as a GeoJSON coordinates array with a single member
	Coordinates [][]float64 `json:"coordinates"`
}

var DefaultCallsigns = []string{"Sky Eye", "Thunderhead", "Eagle Eye", "Ghost Eye", "Sky Keeper", "Bandog", "Long Caster", "Galaxy"}

var DefaultPictureRadius = 300 * unit.NauticalMile

const DefaultMarginRadius = 3 * unit.NauticalMile

var DefaultPlaybackSpeed = 1.0
