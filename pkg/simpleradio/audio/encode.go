package audio

import (
	"context"
	"log/slog"

	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"gopkg.in/hraban/opus.v2"
)

// encodeVoice encodes audio from txChan and publishes an entire transmission's worth of voice packets to packetCh.
func (c *audioClient) encodeVoice(ctx context.Context, packetCh chan<- []voice.VoicePacket) {
	for {
		select {
		case audio := <-c.txChan:
			slog.Debug("encoding transmission from PCM data", "length", len(audio))
			encoder, err := opus.NewEncoder(sampleRate, channels, pcm.OpusApplicationVoIP)
			if err != nil {
				slog.Error("failed to create Opus encoder", "error", err)
				continue
			}

			txPackets := make([]voice.VoicePacket, 0)
			for i := 0; i < len(audio); i += int(frameSize) {
				var frameAudio []float32
				if i+int(frameSize) < len(audio) {
					frameAudio = audio[i : i+int(frameSize)]
				} else {
					frameAudio = audio[i:]
				}
				slog.Debug("encoding audio frame", "frameSize", len(frameAudio), "index", i)
				// Align audio to Opus frame size
				if len(frameAudio) < int(frameSize) {
					previousSize := len(frameAudio)
					padding := make([]float32, int(frameSize)-len(frameAudio))
					frameAudio = append(frameAudio, padding...)
					slog.Debug("padded audio to match frame size", "previousSize", previousSize, "newSize", len(frameAudio))
				}
				audioBytes, err := c.encode(encoder, frameAudio)
				if err != nil {
					slog.Error("failed to encode audio", "error", err)
					continue
				}

				vp := voice.NewVoicePacket(
					audioBytes,
					[]voice.Frequency{
						{
							Frequency:  c.radio.Frequency,
							Modulation: byte(c.radio.Modulation),
							Encryption: 0,
						},
					},
					100000002,
					c.packetNumber,
					0,
					[]byte(c.guid),
					[]byte(c.guid),
				)
				slog.Debug("encoded voice packet", "packet", vp)
				c.packetNumber++
				// TODO transmission struct with attached text and trace id
				txPackets = append(txPackets, vp)
			}
			packetCh <- txPackets
		case <-ctx.Done():
			slog.Info("stopping voice encoder due to context cancellation")
			return
		}
	}
}
