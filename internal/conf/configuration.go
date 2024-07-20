package conf

import (
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/martinlindhe/unit"
)

// Configuration for the SkyEye application.
type Configuration struct {
	// DCSGRPCAddress is the network address of the DCS-gRPC server (including port)
	DCSGRPCAddress string
	// GRPCConnectionTimeout is the connection timeout for connecting to DCS-gRPC
	GRPCConnectionTimeout time.Duration
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
	// SRSCoalition is the coalition that the bot will act on
	SRSCoalition srs.Coalition
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel whisper.Model
}

const DefaultPictureRadius = 35 * unit.NauticalMile
const DefaultMarginRadius = 3 * unit.NauticalMile