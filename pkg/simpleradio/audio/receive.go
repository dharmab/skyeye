package audio

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// receiver contains the state of the current received transmission on a given radio frequency.
type receiver struct {
	lock sync.RWMutex
	// buffer of received voice packets.
	buffer []voice.VoicePacket
	// origin is the GUID of a client we are currently listening to. We can only listen to one client at a time, and whoever started broadcasting first wins.
	origin types.GUID
	// deadline is extended every time another voice packet is received. When we pass the deadline, the transmission is considered over.
	deadline time.Time
	// packetNumber is the number of the last received voice packet. We only record a packet if its packet number is larger than the last received packet's, and skip any that were dropped or delivered out of order.
	// If we were more ambitious we would reassemble the packets and use Opus's forward error correction to recover from lost packets... too bad!
	packetNumber uint64
}

func (r *receiver) receive(vp *voice.VoicePacket) {
	// Accept the packet if it is either:
	// - the first packet of a new transmission
	isNewTransmission := r.origin == "" && r.packetNumber == 0
	// - a newer packet from the same origin
	isNewerPacket := vp.PacketID > uint64(r.packetNumber)
	isSameOrigin := r.origin == types.GUID(vp.OriginGUID)
	shouldAcceptPacket := isNewTransmission || (isNewerPacket && isSameOrigin)
	if !shouldAcceptPacket {
		return
	}

	if isNewTransmission {
		log.Info().Str("origin", string(vp.OriginGUID)).Msg("receiving transmission")
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	r.buffer = append(r.buffer, *vp)
	r.origin = types.GUID(vp.OriginGUID)
	r.deadline = time.Now().Add(maxRxGap)
	r.packetNumber = vp.PacketID
}

func (r *receiver) hasTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	hasPackets := len(r.buffer) > 0
	isComplete := time.Now().After(r.deadline)
	return hasPackets && isComplete
}

func (r *receiver) isReceivingTransmission() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.deadline.After(time.Now())
}

func (r *receiver) reset() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.buffer = make([]voice.VoicePacket, 0)
	r.origin = ""
	r.deadline = time.Time{}
	r.packetNumber = 0
}

// maxRxGap is a duration after which the receiver will assume the end of a transmission if no packets are received.
// TODO make this configurable.
const maxRxGap = 300 * time.Millisecond

// minRxDuration is the mimimum duration of a transmission to be considered for speech recognition. This reduces
// thrashing due to transmissions too short to contain any useful content.
const minRxDuration = 1 * time.Second // 1s is whisper.cpp's minimum duration, it errors for any samples shorter than this.

// receiveUDP listens for incoming UDP packets and routes them to the appropriate channel.
func (c *audioClient) receiveUDP(ctx context.Context, pingCh chan<- []byte, voiceCh chan<- []byte) {
	for {
		if ctx.Err() != nil {
			if ctx.Err() == context.Canceled {
				log.Info().Msg("stopping SRS packet receiver due to context cancellation")
			} else {
				log.Error().Err(ctx.Err()).Msg("stopping packet receiver due to context error")
			}
			return
		}

		udpPacketBuf := make([]byte, 1500)
		n, err := c.connection.Read(udpPacketBuf)
		udpPacket := make([]byte, n)
		copy(udpPacket, udpPacketBuf[0:n])

		switch {
		case err == io.EOF:
			log.Error().Err(err).Msg("UDP connection closed")
		case err != nil:
			log.Error().Err(err).Msg("UDP connection read error")
		case n == 0:
			log.Warn().Err(err).Msg("0 bytes read from UDP connection")
		case n < types.GUIDLength:
			log.Debug().Int("bytes", n).Msg("UDP packet smaller than expected")
		case n == types.GUIDLength:
			// Ping packet
			pingCh <- udpPacket
		case n > types.GUIDLength:
			// Voice packet
			voiceCh <- udpPacket
		}
	}
}

// receivePings listens for incoming UDP ping packets and logs them at DEBUG level.
func (c *audioClient) receivePings(ctx context.Context, in <-chan []byte) {
	for {
		select {
		case b := <-in:
			n := len(b)
			if n < types.GUIDLength {
				log.Debug().Int("bytes", n).Msg("received UDP ping smaller than expected")
			} else if n > types.GUIDLength {
				log.Debug().Int("bytes", n).Msg("received UDP ping larger than expected")
			} else {
				log.Trace().Str("GUID", string(b[0:types.GUIDLength])).Msg("received UDP ping")
				c.lastPing = time.Now()
			}
		case <-ctx.Done():
			log.Info().Msg("stopping SRS ping receiver due to context cancellation")
			return
		}
	}
}

// receiveVoice listens for incoming UDP voice packets, decodes them into VoicePacket structs, and routes them to the out channel for audio decoding.
func (c *audioClient) receiveVoice(ctx context.Context, in <-chan []byte, out chan<- []voice.VoicePacket) {
	// t is a ticker which triggers the check for the end of a transmission.
	t := time.NewTicker(frameLength)
	for {
		select {
		case b := <-in:
			vp, err := decodeVoicePacket(b)
			if err != nil {
				log.Debug().Err(err).Msg("failed to decode voice packet")
				continue
			}
			if vp == nil {
				log.Warn().Msg("nil pointer returned from decodeVoicePacket")
				continue
			}
			for radio, receiver := range c.receivers {
				for _, packetFrequency := range vp.Frequencies {
					testRadio := types.Radio{
						Frequency:   packetFrequency.Frequency,
						Modulation:  types.Modulation(packetFrequency.Modulation),
						IsEncrypted: packetFrequency.Encryption != 0,
					}
					if testRadio.IsSameFrequency(radio) {
						receiver.receive(vp)
					}
				}
			}

		case <-t.C:
			// Check if everyone has stopped talking.
			if len(in) == 0 {
				for _, receiver := range c.receivers {
					if receiver.hasTransmission() {
						duration := time.Duration(len(receiver.buffer)) * frameLength
						logger := log.With().Stringer("duration", duration).Logger()
						if duration > minRxDuration {
							logger.Info().Msg("received transmission")
							audio := make([]voice.VoicePacket, len(receiver.buffer))
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
