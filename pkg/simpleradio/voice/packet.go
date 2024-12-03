// package voice contains the types used by the SRS audio protocol to send and receive audio data over the network.
package voice

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// Packet is a network packet containing:
// A header segment with packet and segment length headers
// An audio segment containing Opus audio
// A frequency segment containing each frequency the audio is transmitted on
// A fixed segment containing metadata
//
// See SRS source code for packet encoding: https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/Network/UDPVoicePacket.cs
type Packet struct {
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

// Frequency describes an audio transmission channel. This struct is only for use in [Packet]. For client information, use [types.Radio] instead.
// Length: 10 bytes.
type Frequency struct {
	// Frequency is the transmission frequency in Hz.
	// Example: 249.500MHz is encoded as 249500000.0
	//
	// Length: 8 bytes
	Frequency float64
	// Modulation is the transmission modulation mode.
	//
	// Length: 1 byte
	Modulation byte
	// Encryption is the transmission encryption mode.
	//
	// Length: 1 byte
	Encryption byte
}

const (
	// headerSegmentLength is the length of the header segment in bytes.
	headerSegmentLength = 6
	// fixedSegmentLength is the length of the fixed segment in bytes.
	fixedSegmentLength = 58
	// frequencyLength is the length of a Frequency in bytes.
	frequencyLength = 10
)

func NewVoicePacket(audioBytes []byte, frequencies []Frequency, unitID uint32, packetID uint64, hops byte, relay []byte, origin []byte) Packet {
	var audioSegmentLength uint16
	if len(audioBytes) > math.MaxUint16 {
		audioSegmentLength = math.MaxUint16
	} else {
		audioSegmentLength = uint16(len(audioBytes))
	}

	var frequenciesSegmentLength uint16
	if len(frequencies)*frequencyLength > math.MaxUint16 {
		frequenciesSegmentLength = math.MaxUint16
	} else {
		frequenciesSegmentLength = uint16(len(frequencies) * frequencyLength)
	}

	return Packet{
		PacketLength:             headerSegmentLength + audioSegmentLength + frequenciesSegmentLength + fixedSegmentLength,
		AudioSegmentLength:       audioSegmentLength,
		FrequenciesSegmentLength: frequenciesSegmentLength,
		AudioBytes:               audioBytes,
		Frequencies:              frequencies,
		UnitID:                   unitID,
		PacketID:                 packetID,
		Hops:                     hops,
		RelayGUID:                relay,
		OriginGUID:               origin,
	}
}

// Encode serializes a VoicePacket into a byte array.
func (p *Packet) Encode() []byte {
	b := make([]byte, p.PacketLength)

	/* Header Segment */
	binary.LittleEndian.PutUint16(b[0:2], p.PacketLength)
	binary.LittleEndian.PutUint16(b[2:4], p.AudioSegmentLength)
	binary.LittleEndian.PutUint16(b[4:6], p.FrequenciesSegmentLength)

	/* Audio Segment */
	copy(b[headerSegmentLength:headerSegmentLength+len(p.AudioBytes)], p.AudioBytes)

	/* Frequencies Segment */
	for i, frequency := range p.Frequencies {
		offset := headerSegmentLength + int(p.AudioSegmentLength) + i*frequencyLength
		binary.LittleEndian.PutUint64(b[offset:offset+8], math.Float64bits(frequency.Frequency))
		b[offset+8] = frequency.Modulation
		b[offset+9] = frequency.Encryption
	}

	/* Fixed Segment */
	fixedSegmentPtr := p.PacketLength - fixedSegmentLength + 1
	unitIDPtr := fixedSegmentPtr
	packetIDPtr := unitIDPtr + 4
	binary.LittleEndian.PutUint32(b[unitIDPtr:packetIDPtr], p.UnitID)

	hopsPtr := packetIDPtr + 8
	binary.LittleEndian.PutUint64(b[packetIDPtr:hopsPtr], p.PacketID)
	b[hopsPtr] = p.Hops

	relayIDPtr := hopsPtr + 1
	originIDPtr := relayIDPtr + types.GUIDLength
	copy(b[relayIDPtr:originIDPtr], p.RelayGUID)
	copy(b[originIDPtr:p.PacketLength], p.OriginGUID)

	return b
}

var _ fmt.Stringer = &Packet{}

func (p *Packet) String() string {
	return fmt.Sprintf(
		"VoicePacket{PacketLength: %d, AudioSegmentLength: %d, FrequenciesSegmentLength: %d, UnitID: %d, PacketID: %d, Hops: %d, RelayGUID: %s, OriginGUID: %s, Frequencies: %v}",
		p.PacketLength,
		p.AudioSegmentLength,
		p.FrequenciesSegmentLength,
		p.UnitID,
		p.PacketID,
		p.Hops,
		p.RelayGUID,
		p.OriginGUID,
		p.Frequencies,
	)
}

// Decode deserializes a voice packet from bytes to struct.
func Decode(b []byte) (packet *Packet, err error) {
	defer func() {
		if r := recover(); r != nil {
			packet = nil
			err = fmt.Errorf("failed to decode VoicePacket: %v", r)
			return
		}
	}()
	// The packet length is the first 2 bytes of the packet.
	packetLength := binary.LittleEndian.Uint16(b[0:2])

	// The fixed segment is at the end of the packet, and each field has a well-known length.
	// Therefore, we can easily decode the fixed segment by working backwards from the end of the packet.
	originIDPtr := packetLength - types.GUIDLength
	relayIDPtr := originIDPtr - types.GUIDLength
	hopsPtr := relayIDPtr - 1
	packetIDPtr := hopsPtr - 8
	unitIDPtr := packetIDPtr - 4

	// Store the packet headers and fixed segment in a VoicePacket struct.
	packet = &Packet{
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

	/* Audio Segment */
	// The audio segment is the next segment after the headers. It always starts at byte 6 and is AudioSegmentLength bytes long.
	audioSegmentPtr := headerSegmentLength
	audioSegment := b[audioSegmentPtr : audioSegmentPtr+int(packet.AudioSegmentLength)]
	packet.AudioBytes = make([]byte, len(audioSegment))
	copy(packet.AudioBytes, audioSegment)

	/* Frequencies Segment */
	// The frequencies segment is the next segment after the audio segment. It always starts at byte 6+AudioSegmentLength and is FrequenciesSegmentLength bytes long.
	frequenciesSegmentPtr := int(6 + packet.AudioSegmentLength)
	frequenciesSegment := b[frequenciesSegmentPtr : frequenciesSegmentPtr+int(packet.FrequenciesSegmentLength)]
	// Iterate over the frequencies segment and decode each frequency.
	for i := 0; i < len(frequenciesSegment); i = i + frequencyLength {
		modulationPtr := i + 8
		encryptionPtr := modulationPtr + 1
		frequency := Frequency{
			Frequency: math.Float64frombits(
				binary.LittleEndian.Uint64(frequenciesSegment[i:modulationPtr]),
			),
			Modulation: frequenciesSegment[modulationPtr],
			Encryption: frequenciesSegment[encryptionPtr],
		}
		packet.Frequencies = append(packet.Frequencies, frequency)
	}

	// That wasn't so bad, was it?
	return
}
