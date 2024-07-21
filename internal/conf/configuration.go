package conf

import (
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
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
	// SRSFrequency is the radio frequency the bot will listen to and talk on in Hz
	SRSFrequency float64
	// Callsign is the GCI callsign used on SRS
	Callsign string
	// Coalition is the coalition that the bot will act on
	Coalition coalitions.Coalition
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel whisper.Model
}

const DefaultPictureRadius = 35 * unit.NauticalMile
const DefaultMarginRadius = 3 * unit.NauticalMile
