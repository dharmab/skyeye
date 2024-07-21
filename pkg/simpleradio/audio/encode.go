package audio

import (
	"context"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
	"gopkg.in/hraban/opus.v2"
)

// Mirror of OPUS_APPLICATION_VOIP from the Opus API
const opusApplicationVoIP = 2048

// encodeVoice encodes audio from txChan and publishes an entire transmission's worth of voice packets to packetCh.
func (c *audioClient) encodeVoice(ctx context.Context, packetCh chan<- []voice.VoicePacket) {
	for {
		select {
		case audio := <-c.txChan:
			log.Debug().Msg("encoding transmission from PCM data")
			encoder, err := opus.NewEncoder(sampleRate, channels, opusApplicationVoIP)
			if err != nil {
				log.Error().Err(err).Msg("failed to create Opus encoder")
				continue
			}

			txPackets := make([]voice.VoicePacket, 0)
			for i := 0; i < len(audio); i += int(frameSize) {
				logger := log.With().Int("index", i).Logger()
				logger.Trace().Msg("encoding audio frame")
				var frameAudio []float32
				if i+int(frameSize) < len(audio) {
					logger.Trace().Msg("encoding full frame")
					frameAudio = audio[i : i+int(frameSize)]
				} else {
					logger.Trace().Msg("encoding partial frame")
					frameAudio = audio[i:]
				}
				logger.Trace().Int("frameSize", len(frameAudio)).Msg("encoding audio frame")
				// Align audio to Opus frame size
				if len(frameAudio) < int(frameSize) {
					previousSize := len(frameAudio)
					logger.Trace().Int("size", previousSize).Msg("data is smaller than frame size")
					padding := make([]float32, int(frameSize)-len(frameAudio))
					frameAudio = append(frameAudio, padding...)
					logger.Trace().Int("previousSize", previousSize).Int("size", len(frameAudio)).Msg("padded audio to match frame size")
				}
				audioBytes, err := c.encode(encoder, frameAudio)
				if err != nil {
					logger.Error().Err(err).Msg("failed to encode audio")
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
				logger.Trace().Interface("packet", vp).Msg("encoded voice packet")
				c.packetNumber++
				// TODO transmission struct with attached text and trace id
				txPackets = append(txPackets, vp)
			}
			log.Debug().Int("count", len(txPackets)).Msg("encoded transmission packets")
			packetCh <- txPackets
		case <-ctx.Done():
			log.Info().Msg("stopping voice encoder due to context cancellation")
			return
		}
	}
}
