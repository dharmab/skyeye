package conf

import (
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
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
	// SRSFrequency is the radio frequency the bot will listen to and talk on in Hz
	SRSFrequency unit.Frequency
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
}

var DefaultCallsigns = []string{"Sky Eye", "Thunderhead", "Eagle Eye", "Ghost Eye", "Sky Keeper", "Bandog", "Long Caster", "Galaxy"}

var DefaultPictureRadius = 300 * unit.NauticalMile

const DefaultMarginRadius = 3 * unit.NauticalMile

var InitialTime time.Time = time.Date(1903, time.December, 17, 2, 35, 0, 0, time.UTC) // https://www.nps.gov/articles/firstflight.htm
