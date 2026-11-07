package simpleradio

import (
	"context"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testGracePeriod = 300 * time.Millisecond

// testGUID pads a readable name out to the length of a real SRS GUID.
func testGUID(name string) types.GUID {
	return types.GUID(name + strings.Repeat("0", types.GUIDLength-len(name)))
}

func testFrequency(radio types.Radio) voice.Frequency {
	return voice.Frequency{Frequency: radio.Frequency, Modulation: byte(radio.Modulation)}
}

// testPacket builds a voice packet carrying a stand-in for an Opus frame. The receiver never looks
// at the audio, only at the identity and frequency metadata.
func testPacket(origin types.GUID, packetID uint64, frequencies ...voice.Frequency) *voice.Packet {
	if len(frequencies) == 0 {
		frequencies = []voice.Frequency{testFrequency(testRadio)}
	}
	packet := voice.NewPacket(
		[]byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		frequencies,
		100000002,
		packetID,
		0,
		[]byte(origin),
		[]byte(origin),
	)
	return &packet
}

// transmit feeds a whole transmission into the receiver, one 40ms frame at a time.
func transmit(r *receiver, origin types.GUID, firstPacketID uint64, frames int) {
	for i := range frames {
		r.receive(testPacket(origin, firstPacketID+uint64(i)))
		time.Sleep(frameLength)
	}
}

func TestReceiverOpensWindowOnFirstPacket(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		r := newReceiver(testGracePeriod)
		assert.False(t, r.isReceivingTransmission())
		assert.True(t, r.deadlineAt().IsZero())

		r.receive(testPacket(testGUID("alpha"), 1))

		assert.True(t, r.isReceivingTransmission(), "window should be open")
		assert.False(t, r.hasTransmission(), "transmission is not finished yet")
		assert.Equal(t, time.Now().Add(testGracePeriod), r.deadlineAt())
	})
}

func TestReceiverAccumulatesTransmission(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		r := newReceiver(testGracePeriod)
		transmit(r, testGUID("alpha"), 1, 25)

		time.Sleep(testGracePeriod)
		packets, audio, span := r.completedTransmission()

		require.Len(t, packets, 25)
		assert.Equal(t, time.Second, audio)
		assert.Equal(t, time.Second, span)
		assert.False(t, r.isReceivingTransmission(), "window should be closed after taking it")
	})
}

// The grace period is the whole boundary between one transmission and the next, so pin both sides
// of it exactly.
func TestReceiverGracePeriodBoundary(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		r := newReceiver(testGracePeriod)
		r.receive(testPacket(testGUID("alpha"), 1))

		time.Sleep(testGracePeriod - time.Nanosecond)
		assert.False(t, r.hasTransmission(), "still within the grace period")

		time.Sleep(2 * time.Nanosecond)
		assert.True(t, r.hasTransmission(), "grace period has elapsed")
	})
}

func TestReceiverSkipsDuplicateAndOutOfOrderPackets(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha := testGUID("alpha")
		r := newReceiver(testGracePeriod)

		r.receive(testPacket(alpha, 5))
		r.receive(testPacket(alpha, 5)) // duplicate
		r.receive(testPacket(alpha, 4)) // delivered out of order
		r.receive(testPacket(alpha, 6))

		time.Sleep(testGracePeriod + time.Nanosecond)
		packets, _, _ := r.completedTransmission()

		require.Len(t, packets, 2)
		assert.Equal(t, uint64(5), packets[0].PacketID)
		assert.Equal(t, uint64(6), packets[1].PacketID)
	})
}

// Radio capture effect: whoever gets the channel first keeps it for the whole transmission.
func TestReceiverFirstTransmitterWinsTheWindow(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha, bravo := testGUID("alpha"), testGUID("bravo")
		r := newReceiver(testGracePeriod)

		r.receive(testPacket(alpha, 1))
		time.Sleep(frameLength)
		r.receive(testPacket(bravo, 100)) // steps on alpha
		time.Sleep(frameLength)
		r.receive(testPacket(alpha, 2))

		time.Sleep(testGracePeriod + time.Nanosecond)
		packets, _, _ := r.completedTransmission()

		require.Len(t, packets, 2)
		for _, packet := range packets {
			assert.Equal(t, alpha, types.GUID(packet.OriginGUID))
		}
	})
}

// Regression: a finished but not yet published transmission used to lock out the next caller,
// because "in progress" meant "state has not been reset" rather than "the deadline has not passed".
func TestReceiverNewTransmitterOpensWindowAfterGracePeriod(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha, bravo := testGUID("alpha"), testGUID("bravo")
		r := newReceiver(testGracePeriod)

		transmit(r, alpha, 1, 3)
		time.Sleep(testGracePeriod + time.Nanosecond)
		first, _, _ := r.completedTransmission()
		require.Len(t, first, 3)

		// bravo has a lower packet number than alpha ever reached, which must not matter.
		r.receive(testPacket(bravo, 1))
		assert.True(t, r.isReceivingTransmission(), "bravo should own a fresh window")

		time.Sleep(testGracePeriod + time.Nanosecond)
		second, _, _ := r.completedTransmission()
		require.Len(t, second, 1)
		assert.Equal(t, bravo, types.GUID(second[0].OriginGUID))
	})
}

// Regression: SRS packet numbers climb for the life of a client, so a packet delayed past the end
// of its own transmission used to be mistaken for the start of a new one. It would buffer a stale
// frame, arm a fresh deadline, and lock out the next real caller for that whole window.
func TestReceiverIgnoresStragglerFromPublishedTransmission(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha, bravo := testGUID("alpha"), testGUID("bravo")
		r := newReceiver(testGracePeriod)

		transmit(r, alpha, 1, 5)
		time.Sleep(testGracePeriod + time.Nanosecond)
		published, _, _ := r.completedTransmission()
		require.Len(t, published, 5)

		// A packet from the middle of the transmission we just published finally turns up.
		r.receive(testPacket(alpha, 3))
		assert.False(t, r.isReceivingTransmission(), "straggler must not open a window")

		// The channel is still free for the next caller.
		r.receive(testPacket(bravo, 1))
		assert.True(t, r.isReceivingTransmission())
		assert.Equal(t, bravo, r.origin)
	})
}

// Remembering packet numbers across transmissions must not silence a client whose numbering
// restarts, which would be far worse than the stale packet the memory exists to reject.
func TestReceiverAcceptsRenumberedTransmitter(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha := testGUID("alpha")
		r := newReceiver(testGracePeriod)

		transmit(r, alpha, 5000, 5)
		time.Sleep(testGracePeriod + time.Nanosecond)
		published, _, _ := r.completedTransmission()
		require.Len(t, published, 5)

		// Well past the point where a delayed packet could still be in flight, the same client
		// starts numbering from scratch.
		time.Sleep(time.Minute)
		r.receive(testPacket(alpha, 1))

		assert.True(t, r.isReceivingTransmission(), "a restarted client must still be heard")
		assert.Equal(t, alpha, r.origin)
	})
}

// Lost packets shorten the audio without shortening the time the speaker was talking. The audio
// duration is what whisper.cpp sees and gates the minimum length; the span is what makes a
// discarded transmission diagnosable.
func TestReceiverReportsAudioAndSpanSeparately(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha := testGUID("alpha")
		r := newReceiver(testGracePeriod)

		// 25 frames' worth of talking, but one packet in every five is lost in transit. The first
		// and last frames arrive, so the span still covers the whole utterance.
		for i := range 25 {
			if i%5 != 2 {
				r.receive(testPacket(alpha, uint64(i+1)))
			}
			time.Sleep(frameLength)
		}

		time.Sleep(testGracePeriod)
		packets, audio, span := r.completedTransmission()

		require.Len(t, packets, 20)
		assert.Equal(t, 800*time.Millisecond, audio, "only the frames that arrived are decodable")
		assert.Equal(t, time.Second, span, "the speaker still talked for a second")
		assert.Less(t, audio, minRxDuration, "which is why a real transmission can be discarded")
	})
}

func TestReceiverResetClearsState(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		alpha := testGUID("alpha")
		r := newReceiver(testGracePeriod)
		transmit(r, alpha, 5, 3)

		r.reset()

		assert.False(t, r.isReceivingTransmission())
		assert.False(t, r.hasTransmission())
		assert.Empty(t, r.buffer)
		assert.Empty(t, r.senders, "a reconnect may hand every transmitter a new GUID")

		// After a reset the receiver accepts a packet it would previously have skipped.
		r.receive(testPacket(alpha, 1))
		assert.True(t, r.isReceivingTransmission())
	})
}

func TestNewReceiverDefaultsGracePeriod(t *testing.T) {
	t.Parallel()
	assert.Equal(t, DefaultSplitTransmissionGracePeriod, newReceiver(0).splitTransmissionGracePeriod)
	assert.Equal(t, time.Second, newReceiver(time.Second).splitTransmissionGracePeriod)
}

// newTestClient builds a Client without touching the network. NewClient dials TCP and UDP, which
// would block outside the synctest bubble.
func newTestClient(gracePeriod time.Duration, radios ...types.Radio) *Client {
	receivers := make(map[types.Radio]*receiver, len(radios))
	for _, radio := range radios {
		receivers[radio] = newReceiver(gracePeriod)
	}
	return &Client{
		clientInfo: types.ClientInfo{Coalition: coalitions.Blue},
		clients:    make(map[types.GUID]types.ClientInfo),
		receivers:  receivers,
	}
}

// Regression: publishing a finished transmission used to be gated on the shared inbound queue being
// empty, which let traffic on any frequency - including ones no radio is tuned to - hold another
// channel's transmission indefinitely.
func TestPublishTransmissionsIgnoresInboundQueueDepth(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		c := newTestClient(testGracePeriod, testRadio)
		out := make(chan []voice.Packet, 4)

		for i := range 25 {
			c.handlePacket(testPacket(testGUID("alpha"), uint64(i+1)).Encode())
			time.Sleep(frameLength)
		}
		time.Sleep(testGracePeriod + time.Nanosecond)

		// A backlog on the inbound queue must not influence a receiver's own deadline.
		in := make(chan []byte, 64)
		for i := range 32 {
			in <- testPacket(testGUID("noise"), uint64(i+1), voice.Frequency{Frequency: 500_000_000}).Encode()
		}
		require.NotEmpty(t, in)

		c.publishTransmissions(out)

		require.Len(t, out, 1)
		assert.Len(t, <-out, 25)
	})
}

func TestReceiveVoicePublishesTransmission(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		c := newTestClient(testGracePeriod, testRadio)
		in := make(chan []byte, 64)
		out := make(chan []voice.Packet, 4)
		go c.receiveVoice(ctx, in, out)

		for i := range 25 {
			in <- testPacket(testGUID("alpha"), uint64(i+1)).Encode()
			time.Sleep(frameLength)
		}
		time.Sleep(testGracePeriod + frameLength)
		synctest.Wait()

		require.Len(t, out, 1)
		assert.Len(t, <-out, 25)
	})
}

func TestReceiveVoiceDiscardsTransmissionBelowMinimum(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		c := newTestClient(testGracePeriod, testRadio)
		in := make(chan []byte, 64)
		out := make(chan []voice.Packet, 4)
		go c.receiveVoice(ctx, in, out)

		for i := range 10 {
			in <- testPacket(testGUID("alpha"), uint64(i+1)).Encode()
			time.Sleep(frameLength)
		}
		time.Sleep(testGracePeriod + frameLength)
		synctest.Wait()

		assert.Empty(t, out, "400ms of audio is below whisper.cpp's one second minimum")
	})
}

// Each channel is monitored on its own, the way a GCI listening to several frequencies hears each
// one separately. Traffic on one must not hold another's transmission open.
func TestReceiveVoiceChannelsAreIndependent(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		c := newTestClient(testGracePeriod, testRadio, testOtherRadio)
		in := make(chan []byte, 128)
		out := make(chan []voice.Packet, 4)
		go c.receiveVoice(ctx, in, out)

		// alpha finishes talking on one channel while bravo keeps talking on the other.
		for i := range 25 {
			in <- testPacket(testGUID("alpha"), uint64(i+1)).Encode()
			in <- testPacket(testGUID("bravo"), uint64(i+1), testFrequency(testOtherRadio)).Encode()
			time.Sleep(frameLength)
		}
		for i := range 25 {
			in <- testPacket(testGUID("bravo"), uint64(i+26), testFrequency(testOtherRadio)).Encode()
			time.Sleep(frameLength)
		}
		synctest.Wait()

		require.Len(t, out, 1, "alpha's transmission should publish while bravo is still talking")
		published := <-out
		require.Len(t, published, 25)
		assert.Equal(t, testGUID("alpha"), types.GUID(published[0].OriginGUID))
	})
}

// Regression: a packet listing the same frequency twice used to buffer its first frame twice,
// putting a 40ms stutter at the front of every transmission.
func TestReceiveVoiceBuffersRepeatedFrequencyOnce(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		c := newTestClient(testGracePeriod, testRadio)
		out := make(chan []voice.Packet, 4)

		duplicated := []voice.Frequency{testFrequency(testRadio), testFrequency(testRadio)}
		for i := range 25 {
			c.handlePacket(testPacket(testGUID("alpha"), uint64(i+1), duplicated...).Encode())
			time.Sleep(frameLength)
		}
		time.Sleep(testGracePeriod + time.Nanosecond)
		c.publishTransmissions(out)

		require.Len(t, out, 1)
		assert.Len(t, <-out, 25)
	})
}

func TestReceiveVoiceHonoursConfiguredGracePeriod(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		gracePeriod := time.Second
		c := newTestClient(gracePeriod, testRadio)
		in := make(chan []byte, 64)
		out := make(chan []voice.Packet, 4)
		go c.receiveVoice(ctx, in, out)

		for i := range 25 {
			in <- testPacket(testGUID("alpha"), uint64(i+1)).Encode()
			time.Sleep(frameLength)
		}

		// A gap that would have ended the transmission under the default grace period.
		time.Sleep(DefaultSplitTransmissionGracePeriod + frameLength)
		synctest.Wait()
		require.Empty(t, out, "the configured grace period has not elapsed yet")

		for i := range 25 {
			in <- testPacket(testGUID("alpha"), uint64(i+26)).Encode()
			time.Sleep(frameLength)
		}
		time.Sleep(gracePeriod + frameLength)
		synctest.Wait()

		require.Len(t, out, 1)
		assert.Len(t, <-out, 50, "both halves belong to one transmission")
	})
}
