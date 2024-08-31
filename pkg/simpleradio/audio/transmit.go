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
			if !c.mute {
				c.tx(packets)
			}

			// Pause between transmissions to sound more natural.
			pause := time.Duration(500+rand.Intn(500)) * time.Millisecond
			time.Sleep(pause)
		case <-ctx.Done():
			log.Info().Msg("stopping SRS audio transmitter due to context cancellation")
			return
		}
	}
}

func (c *audioClient) tx(packets []voice.VoicePacket) {
	for c.lastRx.deadline.After(time.Now()) {
		delay := time.Until(c.lastRx.deadline) + 250*time.Millisecond
		log.Info().Stringer("delay", delay).Msg("delaying outgoing transmission to avoid interrupting incoming transmission")
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
		_, err := c.connection.Write(b)
		if err != nil {
			log.Error().Err(err).Msg("failed to transmit voice packet")
		}
	}
}
