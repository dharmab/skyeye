package simpleradio

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// receiver buffers incoming transmissions on a single radio frequency.
type receiver struct {
	// radio is the frequency this receiver listens on.
	radio types.Radio
	// lock protects the receiver's state.
	lock sync.RWMutex
	// buffer of received voice packets.
	buffer []voice.Packet
	// origin is the GUID of a client we are currently listening to. We can only listen to one client at a time, and whoever started broadcasting first wins.
	origin types.GUID
	// deadline is extended every time another voice packet is received. When we pass the deadline, the transmission is considered over.
	deadline time.Time
	// packetNumber is the number of the last received voice packet. We only record a packet if its packet number is larger than the last received packet's, and skip any that were dropped or delivered out of order.
	// If we were more ambitious we would reassemble the packets and use Opus's forward error correction to recover from lost packets... too bad!
	packetNumber uint64
}

// Receive returns a channel that receives transmissions over the radio. Each transmission is F32LE PCM audio data.
func (c *Client) Receive() <-chan Transmission {
	return c.rxChan
}

// receive checks if the given packet is part of a new transmission or matches a transmission in progress.
// If either case is true, the packet is buffered into the receiver.
func (r *receiver) receive(packet *voice.Packet) {
	// Accept the packet if it is either:
	// - the first packet of a new transmission
	isNewTransmission := r.origin == "" && r.packetNumber == 0
	// - a newer packet from the same origin
	isNewerPacket := packet.PacketID > r.packetNumber
	isSameOrigin := r.origin == types.GUID(packet.OriginGUID)
	shouldAcceptPacket := isNewTransmission || (isNewerPacket && isSameOrigin)
	if !shouldAcceptPacket {
		return
	}

	if isNewTransmission {
		log.Info().Str("origin", string(packet.OriginGUID)).Msg("receiving transmission")
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	if len(r.buffer) < maxRxPackets {
		r.buffer = append(r.buffer, *packet)
	}
	r.origin = types.GUID(packet.OriginGUID)
	r.deadline = time.Now().Add(maxRxGap)
	r.packetNumber = packet.PacketID
}

// isFollowing checks if the receiver is buffering a transmission from the given origin.
func (r *receiver) isFollowing(origin types.GUID) bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.origin == origin
}

// isIdle checks if the receiver is not buffering any transmission.
func (r *receiver) isIdle() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.origin == "" && r.packetNumber == 0
}

// hasTransmission checks if the receiver has a complete transmission buffered.
func (r *receiver) hasTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	hasPackets := len(r.buffer) > 0
	isComplete := time.Now().After(r.deadline)
	return hasPackets && isComplete
}

// isReceivingTransmission checks if the receiver is currently buffering an in-progress transmission.
func (r *receiver) isReceivingTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.deadline.After(time.Now())
}

// reset clears the receiver's buffer.
func (r *receiver) reset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.buffer = make([]voice.Packet, 0)
	r.origin = ""
	r.deadline = time.Time{}
	r.packetNumber = 0
}

// maxRxGap is a duration after which the receiver will assume the end of a transmission if no packets are received.
const maxRxGap = 300 * time.Millisecond

// MaxTransmissionDuration is the longest acceptable received transmission
// length. Longer transmissions are truncated to this.
const MaxTransmissionDuration = 30 * time.Second

// maxRxPackets is the longest acceptable received transmission length in packets.
const maxRxPackets = int(MaxTransmissionDuration / frameLength)

// minRxDuration is the minimum duration of a transmission to be considered for speech recognition. This reduces
// thrashing due to transmissions too short to contain any useful content.
const minRxDuration = 1 * time.Second // 1s is whisper.cpp's minimum duration, it errors for any samples shorter than this.

// routeVoicePacket buffers a voice packet into a single receiver tuned to one of the packet's frequencies.
//
// A transmitter often transmits on several of our frequencies at once (SkyEye itself transmits on all of them). The
// packet must only be buffered once, otherwise the same transmission is decoded, logged and recognized once per
// frequency. The packet goes to the receiver already buffering a transmission from the same origin, or else to the
// first idle receiver on a matching frequency, in configuration order.
func (c *Client) routeVoicePacket(packet *voice.Packet) {
	origin := types.GUID(packet.OriginGUID)
	var target *receiver
	for _, receiver := range c.receivers {
		if !isTunedTo(receiver.radio, packet.Frequencies) {
			continue
		}
		if receiver.isFollowing(origin) {
			target = receiver
			break
		}
		if target == nil && receiver.isIdle() {
			target = receiver
		}
	}
	if target != nil {
		target.receive(packet)
	}
}

// isTunedTo checks if the radio matches any of the given packet frequencies.
func isTunedTo(radio types.Radio, frequencies []voice.Frequency) bool {
	for _, frequency := range frequencies {
		testRadio := types.Radio{
			Frequency:   frequency.Frequency,
			Modulation:  types.Modulation(frequency.Modulation),
			IsEncrypted: frequency.Encryption != 0,
		}
		if testRadio.IsSameFrequency(radio) {
			return true
		}
	}
	return false
}

// receiveVoice listens for incoming UDP voice packets, decodes them into VoicePacket structs, and routes them to the out channel for audio decoding.
func (c *Client) receiveVoice(ctx context.Context, in <-chan []byte, out chan<- []voice.Packet) {
	// t is a ticker which triggers the check for the end of a transmission.
	t := time.NewTicker(frameLength)
	for {
		select {
		case b := <-in:
			packet, err := voice.Decode(b)
			if err != nil {
				log.Debug().Err(err).Msg("failed to decode voice packet")
				continue
			}

			logger := log.With().Str("GUID", string(packet.OriginGUID)).Logger()

			if c.secureCoalitionRadios.Load() {
				client, ok := c.clients[types.GUID(packet.OriginGUID)]
				if !ok {
					logger.Warn().Msg("ignoring voice packet from unknown client")
					continue
				}
				if client.Coalition != c.clientInfo.Coalition {
					logger.Trace().Msg("ignoring voice packet from different coalition")
					continue
				}
			}

			c.routeVoicePacket(packet)
		case <-t.C:
			// Check if everyone has stopped talking.
			if len(in) == 0 {
				for _, receiver := range c.receivers {
					if receiver.hasTransmission() {
						duration := time.Duration(len(receiver.buffer)) * frameLength
						logger := log.With().Stringer("duration", duration).Logger()
						if duration > minRxDuration {
							if len(receiver.buffer) >= maxRxPackets {
								logger.Warn().Stringer("max", MaxTransmissionDuration).Msg("transmission truncated to maximum duration")
							}
							logger.Info().Msg("received transmission")
							audio := make([]voice.Packet, len(receiver.buffer))
							copy(audio, receiver.buffer)
							out <- audio
						} else {
							logger.Info().Msg("discarding transmission below minimum size")
						}
						receiver.reset()
					}
				}
			}
		case <-ctx.Done():
			log.Info().Msg("stopping SRS audio receiver due to context cancellation")
			return
		}
	}
}
