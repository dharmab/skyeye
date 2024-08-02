package audio

import (
	"context"
	"fmt"

	"gopkg.in/hraban/opus.v2"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// decodeVoicePacket decodes a UDP voice packet message into a VoicePacket struct.
func decodeVoicePacket(b []byte) (p *voice.VoicePacket, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("failed to decode VoicePacket: %v", r)
			return
		}
	}()
	vp := voice.NewVoicePacketFrom(b)
	p = &vp
	return
}

// deocdeVoice decodes incoming voice packets from voicePacketsCh into F32LE PCM audio data published to the client's rxChan.
func (c *audioClient) decodeVoice(ctx context.Context, voicePacketsCh <-chan []voice.VoicePacket) {
	for {
		select {
		case voicePackets := <-voicePacketsCh:
			log.Trace().Int("count", len(voicePackets)).Msg("decoding voice packets")
			decoder, err := opus.NewDecoder(sampleRate, channels)
			if err != nil {
				log.Error().Err(err).Msg("failed to create Opus decoder")
				continue
			}
			txPCM := make([]float32, 0)
			for _, vp := range voicePackets {
				log.Trace().Uint64("packetID", vp.PacketID).Int("len", len(vp.AudioBytes)).Msg("decoding voice packet")
				pcm, err := c.decode(decoder, vp.AudioBytes)
				if err != nil {
					log.Error().Err(err).Msg("failed to decode audio")
				} else {
					log.Trace().Int("len", len(pcm)).Msg("decoded voice packet")
					txPCM = append(txPCM, pcm...)
				}
			}

			log.Trace().Int("len", len(txPCM)).Msg("decoded transmission PCM")

			if len(txPCM) > 0 {
				log.Trace().Msg("publishing audio to receiver channel")
				c.rxchan <- txPCM
			} else {
				log.Debug().Msg("decoded transmission PCM is empty")
			}
		case <-ctx.Done():
			log.Info().Msg("stopping voice decoder due to context cancellation")
			return
		}
	}
}
