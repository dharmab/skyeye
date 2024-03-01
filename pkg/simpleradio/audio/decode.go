package audio

import (
	"context"
	"fmt"
	"log/slog"

	"gopkg.in/hraban/opus.v2"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
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
			decoder, err := opus.NewDecoder(sampleRate, channels)
			if err != nil {
				slog.Error("failed to create Opus decoder", "error", err)
				continue
			}
			txPCM := make([]float32, 0)
			for _, vp := range voicePackets {
				pcm, err := c.decode(decoder, vp.AudioBytes)
				if err != nil {
					slog.Error("failed to decode audio", "error", err)
				} else {
					txPCM = append(txPCM, pcm...)
				}
			}

			slog.Debug("decoded transmission PCM", "len", len(txPCM))

			if len(txPCM) > 0 {
				slog.Info("publishing audio to receiver channel")
				c.rxchan <- txPCM
			}
		case <-ctx.Done():
			slog.Info("stopping voice decoder due to context cancellation")
			return
		}
	}
}
