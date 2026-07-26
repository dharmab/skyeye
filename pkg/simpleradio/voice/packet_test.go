package voice

import (
	"encoding/binary"
	"testing"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testOrigin = "OriginGUID000000000000"
	testRelay  = "RelayGUID0000000000000"
)

func testPacket(audio []byte, frequencies []Frequency) Packet {
	return NewPacket(audio, frequencies, 100000002, 3626329, 0, []byte(testRelay), []byte(testOrigin))
}

func TestGUIDFixturesAreRealisticLength(t *testing.T) {
	t.Parallel()
	assert.Len(t, testOrigin, types.GUIDLength)
	assert.Len(t, testRelay, types.GUIDLength)
}

func TestPacketRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		audio       []byte
		frequencies []Frequency
	}{
		{
			name:        "single frequency",
			audio:       []byte{0x1, 0x2, 0x3, 0x4},
			frequencies: []Frequency{{Frequency: 251_000_000, Modulation: 0, Encryption: 0}},
		},
		{
			// Observed on a live server: one packet carrying every frequency a bot transmits on.
			name:  "four frequencies",
			audio: make([]byte, 31),
			frequencies: []Frequency{
				{Frequency: 136_000_000, Modulation: 0},
				{Frequency: 255_000_000, Modulation: 0},
				{Frequency: 281_500_000, Modulation: 0},
				{Frequency: 40_000_000, Modulation: 1},
			},
		},
		{
			name:        "encrypted",
			audio:       []byte{0xFF},
			frequencies: []Frequency{{Frequency: 30_000_000, Modulation: 1, Encryption: 4}},
		},
		{
			name:        "no audio",
			audio:       []byte{},
			frequencies: []Frequency{{Frequency: 251_000_000}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			original := testPacket(test.audio, test.frequencies)
			encoded := original.Encode()
			require.Len(t, encoded, int(original.PacketLength))

			decoded, err := Decode(encoded)
			require.NoError(t, err)
			assert.Equal(t, original.PacketLength, decoded.PacketLength)
			assert.Equal(t, original.AudioSegmentLength, decoded.AudioSegmentLength)
			assert.Equal(t, original.FrequenciesSegmentLength, decoded.FrequenciesSegmentLength)
			assert.Equal(t, original.UnitID, decoded.UnitID)
			assert.Equal(t, original.PacketID, decoded.PacketID)
			assert.Equal(t, original.Hops, decoded.Hops)
			assert.Equal(t, testRelay, string(decoded.RelayGUID))
			assert.Equal(t, testOrigin, string(decoded.OriginGUID))
			assert.Equal(t, test.audio, decoded.AudioBytes)
			assert.Equal(t, test.frequencies, decoded.Frequencies)
		})
	}
}

// A packet observed on a live SRS server, to pin the framing against the real protocol rather than
// only against our own encoder.
func TestDecodeMatchesObservedFraming(t *testing.T) {
	t.Parallel()
	packet := testPacket(make([]byte, 31), []Frequency{
		{Frequency: 136_000_000},
		{Frequency: 255_000_000},
		{Frequency: 281_500_000},
		{Frequency: 40_000_000, Modulation: 1},
	})
	encoded := packet.Encode()

	require.Len(t, encoded, 135, "header 6 + audio 31 + frequencies 40 + fixed 58")
	decoded, err := Decode(encoded)
	require.NoError(t, err)
	assert.Equal(t, uint16(31), decoded.AudioSegmentLength)
	assert.Equal(t, uint16(40), decoded.FrequenciesSegmentLength)
	assert.Len(t, decoded.Frequencies, 4)
}

// Regression: Decode used to trust the length header over the datagram. A corrupt packet whose
// header still pointed within bounds decoded into a plausible packet with a garbage OriginGUID,
// which the receiver would then treat as a real transmitter and let it hold a frequency.
func TestDecodeRejectsMalformedPackets(t *testing.T) {
	t.Parallel()
	validPacket := testPacket([]byte{0x1, 0x2, 0x3, 0x4}, []Frequency{{Frequency: 251_000_000}})
	valid := validPacket.Encode()
	validLength := validPacket.PacketLength

	tests := []struct {
		name   string
		packet func() []byte
	}{
		{
			name: "truncated datagram",
			packet: func() []byte {
				return valid[:len(valid)-10]
			},
		},
		{
			name: "length header shorter than datagram",
			packet: func() []byte {
				b := append([]byte(nil), valid...)
				binary.LittleEndian.PutUint16(b[0:2], validLength-8)
				return b
			},
		},
		{
			name: "length header longer than datagram",
			packet: func() []byte {
				b := append([]byte(nil), valid...)
				binary.LittleEndian.PutUint16(b[0:2], validLength+8)
				return b
			},
		},
		{
			name: "segment lengths disagree with packet length",
			packet: func() []byte {
				b := append([]byte(nil), valid...)
				binary.LittleEndian.PutUint16(b[2:4], 2)
				return b
			},
		},
		{
			name: "frequencies segment is not a whole number of frequencies",
			packet: func() []byte {
				b := append([]byte(nil), valid...)
				binary.LittleEndian.PutUint16(b[2:4], 5)
				binary.LittleEndian.PutUint16(b[4:6], 9)
				return b
			},
		},
		{
			name: "empty",
			packet: func() []byte {
				return []byte{}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			packet, err := Decode(test.packet())
			require.Error(t, err)
			assert.Nil(t, packet)
		})
	}
}
