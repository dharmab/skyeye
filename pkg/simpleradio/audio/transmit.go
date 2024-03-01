package audio

import (
	"context"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
)

// transmit the voice packets from queued transmissions to the SRS server.
func (c *audioClient) transmit(ctx context.Context, packetCh <-chan []voice.VoicePacket) {
	for {
		select {
		case packets := <-packetCh:
			// TODO in-game subtitles
			for _, vp := range packets {
				b := vp.Encode()
				n, err := c.connection.Write(b)
				if err != nil {
					slog.Error("failed to transmit voice packet", "error", err)
				}
				slog.Debug("transmitted voice packet", "packetID", vp.PacketID, "length", n)
				// sleeping half the frame length somehow fixes a PTT stutter issue (???)
				// This might be a performance issue with my debug build of SRS.
				sleepDuration := frameLength / 2
				time.Sleep(sleepDuration)
			}
		case <-ctx.Done():
			slog.Info("stopping voice transmitter due to context cancellation")
			return
		}
	}
}
