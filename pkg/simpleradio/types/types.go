package types

import (
	"time"
)

// Frequency describes an audio transmission channel. This struct is only for use in [VoicePacket]. For client information, use [Radio] instead.
type Frequency struct {
	// Frequency is the transmission freqeuncy in MHz.
	// Example: 249.500MHz is encoded as 249.5
	Frequency float64
	// Modulation is the transmission modulation mode.
	Modulation byte
	// Encryption is the transmission encryption mode.
	Encryption byte
}

// Values from https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/RadioInformation.cs
const (
	// ModulationAM is Amplitude Modulation
	ModulationAM = 0
	// ModulationFM is Frequency Modulation
	ModulationFM = 1
	// ModulationIntercom is intercom (used for multi-crew)
	ModulationIntercom = 2
	ModulationDisabled = 3
	// ModulationIntercom is HAVE QUICK (https://en.wikipedia.org/wiki/Have_Quick, unused?)
	ModulationHAVEQUICK = 4
	// ModulationSATCOM is satellite voice channels (unused?)
	ModulationSATCOM = 5
	// ModulationMIDS is Multifunction Information Distribution System (datalink digital voice channels)
	ModulationMIDS = 6
	// ModulationSINCGARS is Single Channel Ground and Airborne Radio System (https://en.wikipedia.org/wiki/SINCGARS, unused?)
	ModulationSINCGARS = 7
)

// ClientConfiguration is configuration used to construct the audio and data clients
type ClientConfiguration struct {
	// GUID corresponds to [ClientInfo.GUID]
	GUID string
	// Address is the network address of the SRS server, including port
	Address string
	// ConnectionTimeout is the connection timeout for connecting to the SRS server
	ConnectionTimeout time.Duration
	// ClientName corresponds to [ClientInfo.Name]
	ClientName string
	// ExternalAWACSModePassword is the password for External AWACS Mode
	ExternalAWACSModePassword string
	// Coalition corresponds to [ClientInfo.Coalition]
	Coalition Coalition
	// Frequency is the [Frequency] to listen and talk on
	Frequency Frequency
	// AllowRecording corresponds to [ClientInfo.AllowRecording]
	AllowRecording bool
}

// ClientInfo is information about the client included in the message
type ClientInfo struct {
	// Name is the name that will appear in the client list and in in-game transmissions
	Name string `json:"Name"`
	// GUID is some kind of unique ID for the client (???)
	GUID string `json:"ClientGUID"`
	// Seat is the seat number for multicrew aircraft. For bots, set this to 0.
	Seat int `json:"Seat"`
	// Coalition is the side that the client will act on
	Coalition int `json:"Coalition"`
	// AllowRecording indicates consent to record audio server-side. For bots, this should usually be set to True.
	AllowRecording bool         `json:"AllowRecord"`
	Radios         ClientRadios `json:"radios"`
}

type ClientRadios struct {
	UnitID uint64      `json:"unitId"`
	Unit   string      `json:"unit"`
	Radios []Radio     `json:"radios"`
	IFF    Transponder `json:"Transponder"`
}

type Radio struct {
	// Frequency is the transmission freqeuncy in MHz.
	// Example: 249.500MHz is encoded as 249.5
	Frequency float64 `json:"freq"`
	// Modulation is the transmission modulation mode.
	Modulation byte `json:"modulation"`
	// IsEncryption indicates if the transmission is encrypted.
	IsEncrypted bool `json:"enc"`
	// EncruptionKey is the encryption key used to encrypted transmissions.
	EncryptionKey byte `json:"encKey"`
	// GuardFrequency is a second frequency the client can receive.
	GuardFrequency   float64 `json:"secFreq"`
	ShouldRetransmit bool    `json:"retransmit"`
	Volume           float32 `json:"volume"`
}

type MessageType int

const (
	MessageUpdate MessageType = iota
	MessagePing
	MessageSync
	MessageRadioUpdate
	MessageServerSettings
	MessageClientDisconnect
	MessageVersionMismatch
	MessageExternalAWACSModePassword
	MessageExternalAWACSModeDisconnect
)

type Message struct {
	Version                   string            `json:"Version"`
	Type                      MessageType       `json:"MessageType"`
	Client                    ClientInfo        `json:"Client"`
	Clients                   []ClientInfo      `json:"Clients"`
	ServerSettings            map[string]string `json:"ServerSettings"`
	ExternalAWACSModePassword string            `json:"ExternalAWACSModePassword"`
}

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
type IFFControlMode int

const (
	IFFControlModeCockpit  = 0
	IFFControlModeOverlay  = 1
	IFFControlModeDisabled = 2
)

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
type IFFStatus int

const (
	IFFStatusOff    = 0
	IFFStatusNormal = 1
	IFFStatusIdent  = 2
)

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
type Transponder struct {
	ControlMode IFFControlMode `json:"IFFControlMode"`
	Status      IFFStatus      `json:"IFFStatus"`
	Mode1       int            `json:"mode1"` // Use -1 to disable
	Mode3       int            `json:"mode3"` // Use -1 to disable
	Mode4       bool           `json:"mode4"`
	Mic         int            `json:"mic"`
}

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/Network/SRClient.cs
type Coalition int

const (
	CoalitionRed       = 1
	CoalitionBlue      = 2
	CoalitionSpectator = 3
)
