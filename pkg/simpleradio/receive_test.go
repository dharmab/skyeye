package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/stretchr/testify/assert"
)

var (
	routeRadioA = types.Radio{Frequency: 254_000_000, Modulation: types.ModulationAM}
	routeRadioB = types.Radio{Frequency: 145_000_000, Modulation: types.ModulationAM}
	routeRadioC = types.Radio{Frequency: 30_000_000, Modulation: types.ModulationFM}
)

func newRouteTestClient() *Client {
	return &Client{
		receivers: []*receiver{{radio: routeRadioA}, {radio: routeRadioB}, {radio: routeRadioC}},
	}
}

func routeTestPacket(origin string, id uint64, radios ...types.Radio) *voice.Packet {
	frequencies := make([]voice.Frequency, 0, len(radios))
	for _, radio := range radios {
		frequencies = append(frequencies, voice.Frequency{Frequency: radio.Frequency, Modulation: byte(radio.Modulation)})
	}
	return &voice.Packet{OriginGUID: []byte(origin), PacketID: id, Frequencies: frequencies}
}

func bufferedPackets(c *Client) []int {
	counts := make([]int, 0, len(c.receivers))
	for _, r := range c.receivers {
		counts = append(counts, len(r.buffer))
	}
	return counts
}

func TestRouteVoicePacketMultipleFrequencies(t *testing.T) {
	t.Parallel()
	c := newRouteTestClient()
	for id := uint64(1); id <= 5; id++ {
		c.routeVoicePacket(routeTestPacket("origin0000000000000001", id, routeRadioC, routeRadioB, routeRadioA))
	}
	assert.Equal(t, []int{5, 0, 0}, bufferedPackets(c))
}

func TestRouteVoicePacketSimultaneousTransmitters(t *testing.T) {
	t.Parallel()
	c := newRouteTestClient()
	for id := uint64(1); id <= 3; id++ {
		c.routeVoicePacket(routeTestPacket("origin0000000000000001", id, routeRadioA, routeRadioB))
		c.routeVoicePacket(routeTestPacket("origin0000000000000002", id, routeRadioA, routeRadioB))
	}
	assert.Equal(t, []int{3, 3, 0}, bufferedPackets(c))
	assert.Equal(t, types.GUID("origin0000000000000001"), c.receivers[0].origin)
	assert.Equal(t, types.GUID("origin0000000000000002"), c.receivers[1].origin)
}

func TestRouteVoicePacketFollowsOriginAcrossFrequencyChange(t *testing.T) {
	t.Parallel()
	c := newRouteTestClient()
	c.routeVoicePacket(routeTestPacket("origin0000000000000001", 1, routeRadioB))
	c.routeVoicePacket(routeTestPacket("origin0000000000000001", 2, routeRadioA, routeRadioB))
	assert.Equal(t, []int{0, 2, 0}, bufferedPackets(c))
}

func TestRouteVoicePacketOtherFrequency(t *testing.T) {
	t.Parallel()
	c := newRouteTestClient()
	c.routeVoicePacket(routeTestPacket("origin0000000000000001", 1, types.Radio{Frequency: 251_000_000, Modulation: types.ModulationAM}))
	assert.Equal(t, []int{0, 0, 0}, bufferedPackets(c))
}
