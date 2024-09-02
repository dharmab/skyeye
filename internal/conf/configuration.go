package conf

import (
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/martinlindhe/unit"
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
	// Callsign is the GCI callsign used on SRS
	Callsign string
	// Coalition is the coalition that the bot will act on
	Coalition coalitions.Coalition
	// RadarSweepInterval is the rate at which the radar will update. This does not impact performance - ACMI data is still streamed at the same rate.
	// It only impacts the update rate of the GCI radar picture.
	RadarSweepInterval time.Duration
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel *whisper.Model
	// Voice is the voice used for SRS transmissions
	Voice voices.Voice
	// Mute disables SRS transmissions
	Mute bool
	// Piper playback speed (default is 1.0) - The higher the value the slower it is.
	PlaybackSpeed float32
	// Piper playback pause after every sentence in seconds (default is 0.2)
	PlaybackPause time.Duration
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
}

var DefaultCallsigns = []string{"Sky Eye", "Thunderhead", "Eagle Eye", "Ghost Eye", "Sky Keeper", "Bandog", "Long Caster", "Galaxy"}

var DefaultPictureRadius = 300 * unit.NauticalMile

const DefaultMarginRadius = 3 * unit.NauticalMile

var DefaultPlaybackSpeed = 1.0
