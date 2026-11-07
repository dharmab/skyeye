package simpleradio

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// DefaultSplitTransmissionGracePeriod is used when the client configuration does not set one.
const DefaultSplitTransmissionGracePeriod = 300 * time.Millisecond

// minRxDuration is the minimum duration of a transmission to be considered for speech recognition. This reduces
// thrashing due to transmissions too short to contain any useful content.
const minRxDuration = 1 * time.Second // 1s is whisper.cpp's minimum duration, it errors for any samples shorter than this.

// senderTTL is how long a transmitter's packet high-water mark is remembered after its last packet.
// It only needs to outlast network delays by a wide margin.
const senderTTL = 5 * time.Minute

// sender is the most recent packet seen from one transmitting client.
type sender struct {
	// packetNumber is the highest packet number seen from this transmitter.
	packetNumber uint64
	// at is when that packet was received.
	at time.Time
}

// receiver buffers incoming transmissions on a single radio frequency.
//
// The channel owns the transmission window, not the caller: the window opens when a packet arrives
// on an idle channel and closes when no packet has arrived for splitTransmissionGracePeriod. Each
// receiver's window is independent of every other receiver's, the way a GCI monitoring several
// channels hears each one separately.
//
// Within a window the first transmitter wins and anyone stepping on them is dropped, mirroring
// radio capture effect. That also keeps the buffer single-origin, which the Opus decoder requires.
type receiver struct {
	// lock protects the receiver's state.
	lock sync.RWMutex
	// buffer of received voice packets.
	buffer []voice.Packet
	// origin is the GUID of the client that won the current transmission window.
	origin types.GUID
	// startedAt is when the current transmission window opened.
	startedAt time.Time
	// deadline is extended every time another voice packet is received. Once it passes, the
	// transmission is considered over.
	deadline time.Time
	// splitTransmissionGracePeriod is how long to wait for another packet before treating a
	// transmission as finished.
	splitTransmissionGracePeriod time.Duration
	// senders records the highest packet number seen from each transmitter. It deliberately
	// outlives individual transmissions: SRS packet numbers increment for the life of a client, so
	// remembering them is what stops a packet delayed past the end of its own transmission from
	// opening a new one.
	//
	// If we were more ambitious we would reassemble the packets and use Opus's forward error
	// correction to recover from lost packets... too bad!
	senders map[types.GUID]sender
}

// newReceiver returns a receiver listening on one frequency. A grace period of zero selects
// DefaultSplitTransmissionGracePeriod.
func newReceiver(gracePeriod time.Duration) *receiver {
	if gracePeriod <= 0 {
		gracePeriod = DefaultSplitTransmissionGracePeriod
	}
	return &receiver{
		splitTransmissionGracePeriod: gracePeriod,
		senders:                      make(map[types.GUID]sender),
	}
}

// Receive returns a channel that receives transmissions over the radio. Each transmission is F32LE PCM audio data.
func (c *Client) Receive() <-chan Transmission {
	return c.rxChan
}

// receive buffers the given packet if it continues the transmission in progress or starts a new
// one.
//
// Any completed transmission must be taken with completedTransmission before calling this,
// otherwise a new transmission would be appended onto the previous one's buffer.
func (r *receiver) receive(packet *voice.Packet) {
	now := time.Now()
	origin := types.GUID(packet.OriginGUID)

	r.lock.Lock()
	defer r.lock.Unlock()

	// Skip duplicates, packets delivered out of order, and packets delayed past the end of the
	// transmission they belong to.
	//
	// A packet number that goes backwards only means a stale packet if it turns up while that
	// transmitter is still being heard. Long afterwards it more likely means the transmitter
	// restarted its numbering, and refusing it would silence that client for good.
	if s, ok := r.senders[origin]; ok && packet.PacketID <= s.packetNumber {
		if now.Sub(s.at) <= r.stragglerWindow() {
			return
		}
		log.Debug().
			Str("origin", string(origin)).
			Uint64("packetNumber", packet.PacketID).
			Uint64("previousPacketNumber", s.packetNumber).
			Msg("transmitter appears to have restarted its packet numbering")
	}

	isWindowOpen := len(r.buffer) > 0 && now.Before(r.deadline)
	if isWindowOpen && origin != r.origin {
		// Someone stepped on the client we are already listening to. Whoever started first wins.
		return
	}

	if !isWindowOpen {
		log.Info().Str("origin", string(origin)).Msg("receiving transmission")
		r.origin = origin
		r.startedAt = now
	}

	r.senders[origin] = sender{packetNumber: packet.PacketID, at: now}
	if len(r.buffer) < maxRxPackets {
		r.buffer = append(r.buffer, *packet)
	}
	r.deadline = now.Add(r.splitTransmissionGracePeriod)
}

// completedTransmission returns the buffered transmission if its window has closed, the duration of
// audio it contains, and the wall clock span over which it arrived. It returns nil packets if no
// transmission is ready. Taking a transmission closes the window.
//
// The audio duration and the span differ when packets are lost: the audio duration is what
// whisper.cpp will see, while the span is how long the speaker was actually talking.
func (r *receiver) completedTransmission() (packets []voice.Packet, audio, span time.Duration) {
	now := time.Now()

	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r.buffer) == 0 || !now.After(r.deadline) {
		return nil, 0, 0
	}

	packets = r.buffer
	audio = time.Duration(len(packets)) * frameLength
	lastPacketAt := r.deadline.Add(-r.splitTransmissionGracePeriod)
	span = lastPacketAt.Sub(r.startedAt) + frameLength

	r.buffer = nil
	r.origin = ""
	r.startedAt = time.Time{}
	r.deadline = time.Time{}
	r.forgetStaleSenders(now)

	return packets, audio, span
}

// hasTransmission checks if the receiver has a complete transmission buffered.
func (r *receiver) hasTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return len(r.buffer) > 0 && time.Now().After(r.deadline)
}

// isReceivingTransmission checks if the receiver is currently buffering an in-progress transmission.
func (r *receiver) isReceivingTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.deadline.After(time.Now())
}

// deadlineAt returns the time at which the current transmission window closes. It is the zero time
// if no transmission is in progress.
func (r *receiver) deadlineAt() time.Time {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.deadline
}

// reset discards any transmission in progress. It is called when reconnecting to the SRS server,
// where any partial transmission is no longer useful and every transmitter may have a new GUID.
func (r *receiver) reset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.buffer = nil
	r.origin = ""
	r.startedAt = time.Time{}
	r.deadline = time.Time{}
	clear(r.senders)
}

// stragglerWindow is how long after a transmitter's last packet a lower-numbered packet is still
// treated as one that was delayed rather than one that was renumbered. A packet later than this is
// long past being useful audio anyway: measured against a live server, packets arrive 40ms apart
// with a worst case of 52ms.
func (r *receiver) stragglerWindow() time.Duration {
	return 2 * r.splitTransmissionGracePeriod
}

// MaxTransmissionDuration is the longest acceptable received transmission
// length. Longer transmissions are truncated to this.
const MaxTransmissionDuration = 30 * time.Second

// maxRxPackets is the longest acceptable received transmission length in packets.
const maxRxPackets = int(MaxTransmissionDuration / frameLength)

// forgetStaleSenders drops the high-water marks of transmitters we have not heard from in a while,
// bounding the map on a long-running server. The caller must hold the lock.
func (r *receiver) forgetStaleSenders(now time.Time) {
	for guid, s := range r.senders {
		if now.Sub(s.at) > senderTTL {
			delete(r.senders, guid)
		}
	}
}

// receiveVoice listens for incoming UDP voice packets, decodes them into VoicePacket structs, and routes them to the out channel for audio decoding.
func (c *Client) receiveVoice(ctx context.Context, in <-chan []byte, out chan<- []voice.Packet) {
	// ticker triggers the check for the end of a transmission.
	ticker := time.NewTicker(frameLength)
	for {
		select {
		case b := <-in:
			// Publish any transmission that has already ended before buffering this packet, so
			// that a finished transmission never blocks the next one from starting.
			c.publishTransmissions(out)
			c.handlePacket(b)
		case <-ticker.C:
			// Handle everything already queued before checking deadlines. Otherwise a loop that
			// stalled could declare a transmission over while the packets that would have extended
			// it are still waiting to be read.
			c.drainPackets(in)
			c.publishTransmissions(out)
		case <-ctx.Done():
			log.Info().Msg("stopping SRS audio receiver due to context cancellation")
			return
		}
	}
}

// drainPackets handles every packet currently queued, without blocking.
func (c *Client) drainPackets(in <-chan []byte) {
	for {
		select {
		case b := <-in:
			c.handlePacket(b)
		default:
			return
		}
	}
}

// handlePacket decodes a voice packet and buffers it into each receiver tuned to a frequency the
// packet was transmitted on.
func (c *Client) handlePacket(b []byte) {
	packet, err := voice.Decode(b)
	if err != nil {
		log.Debug().Err(err).Msg("failed to decode voice packet")
		return
	}

	logger := log.With().Str("GUID", string(packet.OriginGUID)).Logger()

	if c.secureCoalitionRadios.Load() {
		peer, ok := c.getPeer(types.GUID(packet.OriginGUID))
		if !ok {
			logger.Warn().Msg("ignoring voice packet from unknown client")
			return
		}
		if peer.Coalition != c.clientInfo.Coalition {
			logger.Trace().Msg("ignoring voice packet from different coalition")
			return
		}
	}

	for radio, receiver := range c.receivers {
		for _, frequency := range packet.Frequencies {
			testRadio := types.Radio{
				Frequency:   frequency.Frequency,
				Modulation:  types.Modulation(frequency.Modulation),
				IsEncrypted: frequency.Encryption != 0,
			}
			if testRadio.IsSameFrequency(radio) {
				receiver.receive(packet)
				// A packet can list the same frequency more than once. The per-transmitter packet
				// numbering would skip the repeat anyway; stopping here saves the second lookup.
				break
			}
		}
	}
}

// publishTransmissions publishes each transmission whose window has closed. Every receiver's window
// is evaluated on its own, so a busy channel cannot hold another channel's transmission open.
func (c *Client) publishTransmissions(out chan<- []voice.Packet) {
	for _, receiver := range c.receivers {
		packets, audio, span := receiver.completedTransmission()
		if packets == nil {
			continue
		}
		logger := log.With().
			Stringer("duration", audio).
			Stringer("span", span).
			Int("packets", len(packets)).
			Logger()
		if audio >= minRxDuration {
			if len(packets) >= maxRxPackets {
				logger.Warn().Stringer("max", MaxTransmissionDuration).Msg("transmission truncated to maximum duration")
			}
			logger.Info().Msg("received transmission")
			out <- packets
		} else {
			// A span much longer than the duration means packets were lost, rather than the
			// speaker simply saying something short.
			logger.Info().Msg("discarding transmission below minimum size")
		}
	}
}
