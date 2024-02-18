package audio

import (
	"encoding/binary"
	"math"
)

// VoicePacket is a network packet containing:
// A header segment with packet and segment length headers
// An audio segment containing Opus audio
// A frequency segment containing each frequency the audio is transmitted on
// A fixed segment containing metadata
//
// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/Network/UDPVoicePacket.cs
type VoicePacket struct {
	/* Headers */
	// PacketLength is the total packet length in bytes.
	PacketLength uint16
	// AudioSegmentLength is the length of the Audio segment struct. This is not the length of the audio itself!
	AudioSegmentLength uint16
	// FrequenciesSegmentLength is the length of the Frequencies segment.
	FrequenciesSegmentLength uint16

	/* Audio segment */
	// AudioLength is the length of the AudioPart1 byte array.
	AudioLength uint16
	// AudioBytes is the AudioPart1 byte array. There is no Part2.
	AudioBytes []byte

	/* Frequencies Segment */
	// Frequencies is an array of information for each frequency+modulation+encryption combination the audio is transmitted on.
	Frequencies []Frequency

	/* Fixed Segment */
	// UnitID is the ID of the in-game unit (?)
	UnitID uint32
	// PacketID is the ID of this packet
	PacketID uint64
	// RetransmissionCount is the number of retransmissions. It is used to detect excessive retries.
	RetransmissionCount byte
	// 22 bytes ASCII string
	OriginalGUID []byte
	// 22 bytes ASCII string
	GUID []byte
}

// Frequency describes an audio transmission channel. This struct is only for use in [VoicePacket]. For client information, use [srs.Radio] instead.
type Frequency struct {
	// Frequency is the transmission frequency in Hz.
	// Example: 249.500MHz is encoded as 249500000.0
	Frequency float64
	// Modulation is the transmission modulation mode.
	Modulation byte
	// Encryption is the transmission encryption mode.
	Encryption byte
}

// newVoicePacketFrom converts a voice packet from bytes to struct
func newVoicePacketFrom(b []byte) VoicePacket {
	p := VoicePacket{
		PacketLength:             binary.LittleEndian.Uint16(b[0:2]),
		AudioSegmentLength:       binary.LittleEndian.Uint16(b[2:4]),
		FrequenciesSegmentLength: binary.LittleEndian.Uint16(b[4:6]),
	}
	// slog.Debug("packet", "struct", p)

	audioSegmentOffset := 6
	audioBytesOffset := audioSegmentOffset + 2
	p.AudioLength = binary.LittleEndian.Uint16(b[audioSegmentOffset:audioBytesOffset])
	frequenciesOffset := audioBytesOffset + int(p.AudioLength)
	p.AudioBytes = b[audioBytesOffset:frequenciesOffset]
	fixedSegmentOffset := frequenciesOffset + int(p.FrequenciesSegmentLength)

	for i := frequenciesOffset; i <= frequenciesOffset+int(p.FrequenciesSegmentLength); i = i + 110 {
		frequency := Frequency{
			Frequency:  math.Float64frombits(binary.LittleEndian.Uint64(b[i : i+8])),
			Modulation: b[i+9],
			Encryption: b[i+10],
		}

		p.Frequencies = append(p.Frequencies, frequency)
	}

	unitIDOffset := fixedSegmentOffset
	packetIDOffset := unitIDOffset + 4
	retrasmissionCountOffset := packetIDOffset + 8
	originalGUIDOffset := retrasmissionCountOffset + 1
	guidOffset := originalGUIDOffset + 22

	p.UnitID = binary.LittleEndian.Uint32(b[unitIDOffset:packetIDOffset])
	p.PacketID = binary.LittleEndian.Uint64(b[packetIDOffset:retrasmissionCountOffset])
	p.RetransmissionCount = b[retrasmissionCountOffset]
	p.OriginalGUID = b[originalGUIDOffset:guidOffset]
	p.GUID = b[guidOffset : guidOffset+22]

	return p
}
