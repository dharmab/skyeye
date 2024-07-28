package audio

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// transmit the voice packets from queued transmissions to the SRS server.
func (c *audioClient) transmit(ctx context.Context, packetCh <-chan []voice.VoicePacket) {
	for {
		select {
		case packets := <-packetCh:
			c.tx(ctx, packets)
		case <-ctx.Done():
			log.Info().Msg("stopping voice transmitter due to context cancellation")
			return
		}
	}
}

func (c *audioClient) tx(ctx context.Context, packets []voice.VoicePacket) {
	c.busy.Lock()
	defer c.busy.Unlock()
	// TODO in-game subtitles
	for _, vp := range packets {
		startTime := time.Now()
		b := vp.Encode()
		n, err := c.connection.Write(b)
		if err != nil {
			log.Error().Err(err).Msg("failed to transmit voice packet")
		} else {
			log.Trace().Uint64("packetID", vp.PacketID).Int("length", n).Msg("transmitted voice packet")

		}
		// sleeping half the frame length somehow fixes a PTT stutter issue (???)
		// This might be a performance issue with my debug build of SRS.
		sleepDuration := (frameLength - (time.Since(startTime) * 512))
		time.Sleep(sleepDuration)
	}
}
