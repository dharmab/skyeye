package types

import (
	"math"
)

// This file implements types from https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/RadioInformation.cs

// Modulation indicates the technology used to send a transmission.
type Modulation byte

const (
	// ModulationAM is Amplitude Modulation
	ModulationAM = 0
	// ModulationFM is Frequency Modulation
	ModulationFM = 1
	// ModulationIntercom is intercom (used for multi-crew)
	ModulationIntercom = 2
	ModulationDisabled = 3
	// ModulationIntercom is HAVE QUICK (https://en.wikipedia.org/wiki/Have_Quick, unused)
	ModulationHAVEQUICK = 4
	// ModulationSATCOM is satellite voice channels (unused)
	ModulationSATCOM = 5
	// ModulationMIDS is Multifunction Information Distribution System (datalink digital voice channels)
	// These are used by F/A-18C for VOC A and VOC B
	ModulationMIDS = 6
	// ModulationSINCGARS is Single Channel Ground and Airborne Radio System (https://en.wikipedia.org/wiki/SINCGARS, unused)
	ModulationSINCGARS = 7
)

// Radio describes one of a client's radios.
type Radio struct {
	// Frequency is the transmission frequency in Hz.
	// Example: 249.500MHz is encoded as 249500000.0
	Frequency float64 `json:"freq"`
	// Modulation is the transmission modulation mode.
	Modulation Modulation `json:"modulation"`
	// IsEncryption indicates if the transmission is encrypted.
	IsEncrypted bool `json:"enc"`
	// EncruptionKey is the encryption key used to encrypted transmissions.
	EncryptionKey byte `json:"encKey"`
	// GuardFrequency is a second frequency the client can receive.
	GuardFrequency   float64 `json:"secFreq"`
	ShouldRetransmit bool    `json:"retransmit"`
}

// IsSameFrequency is true if the other radio has the same frequency, modulation, and encryption settings as this radio.
func (r Radio) IsSameFrequency(other Radio) bool {
	// 1KHz range acceptable
	doesFrequencyMatch := math.Abs(float64(r.Frequency)-float64(other.Frequency)) <= 500.0
	doesModulationMatch := r.Modulation == other.Modulation
	doesEncryptionMatch := (!r.IsEncrypted && !other.IsEncrypted) || (r.IsEncrypted && other.IsEncrypted && r.EncryptionKey == other.EncryptionKey)
	return doesFrequencyMatch && doesModulationMatch && doesEncryptionMatch
}
