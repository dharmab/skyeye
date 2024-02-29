// package voice contains the types used by the SRS audio protocol to send and receive audio data over the network.
package voice

import (
	"encoding/binary"
	"log/slog"
	"math"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// VoicePacket is a network packet containing:
// A header segment with packet and segment length headers
// An audio segment containing Opus audio
// A frequency segment containing each frequency the audio is transmitted on
// A fixed segment containing metadata
//
// See SRS source code for packet encoding: https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/Network/UDPVoicePacket.cs
type VoicePacket struct {
	/* Headers */

	// PacketLength is the total packet length in bytes.
	//
	// Bytes: 0:2
	//
	// Length: 2 bytes
	PacketLength uint16
	// AudioSegmentLength is the length of the Audio segment struct.
	//
	// Bytes: 2:4
	//
	// Length: 2 bytes
	AudioSegmentLength uint16
	// FrequenciesSegmentLength is the length of the Frequencies segment.
	//
	// Bytes: 4:6
	//
	// Length: 2 bytes
	FrequenciesSegmentLength uint16

	/* Audio segment */
	// AudioBytes is the AudioPart1 byte array. This is the audio data  as an Opus bitstream, encoded as 16KHz Mono in 40ms frames.
	// The upstream name is directly mirrored from the IDirectSoundBuffer::Lock function in the legacy DirectSound API - Part2 is not used by SRS.
	//
	// Bytes: 6:6+AudioSegmentLength
	//
	// Length: AudioSegmentLength
	AudioBytes []byte

	/* Frequencies Segment */

	// Frequencies is an array of information for each frequency, modulation and encryption combination the audio is transmitted on.
	//
	// Bytes: 6+AudioSegmentLength:6+AudioSegmentLength+FrequenciesSegmentLength
	//
	// Length: FrequenciesSegmentLength
	Frequencies []Frequency

	/* Fixed Segment */

	// UnitID is the ID of the in-game unit that originated the packet.
	//
	// Bytes: PacketLength-58:PacketLength-53
	//
	// Length: 4 bytes
	UnitID uint32
	// PacketID is the ID of this packet. Packets from the same transmitter increment by 1 for each transmission.
	//
	// Bytes: PacketLength-53:PacketLength-45
	//
	// Length: 8 bytes
	PacketID uint64
	// Hops is the number of retransmissions. This value is checked in SRS to limit retransmisisons.
	//
	// Bytes: PacketLength-45:PacketLength-44
	//
	// Length: 1 byte
	Hops byte
	// RelayGUID is the GUID of the last transmitter. This may differ from OriginGUID if this is a retransmission.
	//
	// Bytes: PacketLength-44:PacketLength-22
	//
	// Length: 22 bytes
	RelayGUID []byte
	// OriginGUID is the GUID of the original transmitter.
	//
	// Bytes: PacketLength-22:PacketLength
	//
	// Length: 22 bytes
	OriginGUID []byte
}

// Frequency describes an audio transmission channel. This struct is only for use in [VoicePacket]. For client information, use [types.Radio] instead.
// Length: 10
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
func NewVoicePacketFrom(b []byte) VoicePacket {
	// The packet length is the first 2 bytes of the packet.
	packetLength := binary.LittleEndian.Uint16(b[0:2])
	slog.Debug("decoded voice packet length header", "value", packetLength)

	// The fixed segment is at the end of the packet, and each field has a well-known length.
	// Therefore, we can easily decode the fixed segment by working backwards from the end of the packet.
	originIDPtr := packetLength - types.GUIDLength
	relayIDPtr := originIDPtr - types.GUIDLength
	hopsPtr := relayIDPtr - 1
	packetIDPtr := hopsPtr - 8
	unitIDPtr := packetIDPtr - 4

	// Store the packet headers and fixed segment in a VoicePacket struct.
	packet := VoicePacket{
		/* Headers */
		PacketLength:             packetLength,
		AudioSegmentLength:       binary.LittleEndian.Uint16(b[2:4]),
		FrequenciesSegmentLength: binary.LittleEndian.Uint16(b[4:6]),
		/* Fixed Segment */
		UnitID:     binary.LittleEndian.Uint32(b[unitIDPtr:packetIDPtr]),
		PacketID:   binary.LittleEndian.Uint64(b[packetIDPtr:hopsPtr]),
		Hops:       b[hopsPtr],
		RelayGUID:  b[relayIDPtr : relayIDPtr+types.GUIDLength],
		OriginGUID: b[originIDPtr : originIDPtr+types.GUIDLength],
	}
	slog.Debug("decoded voice packet headers and fixed segment", "struct", packet)

	/* Audio Segment */
	// The audio segment is the next segment after the headers. It always starts at byte 6 and is AudioSegmentLength bytes long.
	audioSegmentPtr := 6
	audioSegment := b[audioSegmentPtr : audioSegmentPtr+int(packet.AudioSegmentLength)]
	packet.AudioBytes = make([]byte, len(audioSegment))
	copy(packet.AudioBytes, audioSegment)
	slog.Debug("decoded voice packet audio bytes", "length", len(packet.AudioBytes))

	/* Frequencies Segment */
	// The frequencies segment is the next segment after the audio segment. It always starts at byte 6+AudioSegmentLength and is FrequenciesSegmentLength bytes long.
	frequenciesSegmentPtr := int(6 + packet.AudioSegmentLength)
	frequenciesSegment := b[frequenciesSegmentPtr : frequenciesSegmentPtr+int(packet.FrequenciesSegmentLength)]
	// Each frequency is 10 bytes long, so we can iterate over the segment in 10 byte chunks to decode each frequency.
	for i := 0; i < len(frequenciesSegment); i = i + 10 {
		modulationPtr := i + 8
		encryptionPtr := modulationPtr + 1
		frequency := Frequency{
			Frequency: math.Float64frombits(
				binary.LittleEndian.Uint64(frequenciesSegment[i : i+8]),
			),
			Modulation: frequenciesSegment[modulationPtr],
			Encryption: frequenciesSegment[encryptionPtr],
		}
		packet.Frequencies = append(packet.Frequencies, frequency)
	}
	slog.Debug("decoded voice packet frequencies segment", "frequencies", packet.Frequencies)

	// That wasn't so bad, was it?

	return packet
}
