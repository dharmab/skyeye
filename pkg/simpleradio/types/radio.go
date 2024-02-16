package types

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
}
