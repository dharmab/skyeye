package audio

import (
	"context"
	"math/rand"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// transmit the voice packets from queued transmissions to the SRS server.
func (c *audioClient) transmit(ctx context.Context, packetCh <-chan []voice.VoicePacket) {
	for {
		select {
		case packets := <-packetCh:
			c.tx(packets)
			// Pause between transmissions to sound more natural.
			pause := time.Duration(250+rand.Intn(400)) * time.Millisecond
			time.Sleep(pause)
		case <-ctx.Done():
			log.Info().Msg("stopping voice transmitter due to context cancellation")
			return
		}
	}
}

func (c *audioClient) tx(packets []voice.VoicePacket) {
	if c.lastRx.deadline.After(time.Now()) {
		delay := 250 * time.Millisecond
		log.Info().Dur("delay", delay).Msg("delaying outgoing transmission to avoid interrupting incoming transmission")
		time.Sleep(delay)
	}
	c.busy.Lock()
	defer c.busy.Unlock()
	// TODO in-game subtitles
	startTime := time.Now()
	for i, vp := range packets {
		b := vp.Encode()
		// Tight timing is important here - don't write the next packet until halfway through the previous packet's frame.
		// Write too quickly, and the server will skip audio to play the latest packet.
		// Write too slowly, and the transmission will stutter.
		delay := time.Until(
			startTime.
				Add(time.Duration(i) * frameLength).
				Add(-frameLength / 2),
		)
		time.Sleep(delay)
		n, err := c.connection.Write(b)
		if err != nil {
			log.Error().Err(err).Msg("failed to transmit voice packet")
		} else {
			log.Trace().Uint64("packetID", vp.PacketID).Int("length", n).Msg("transmitted voice packet")
		}
	}
}
