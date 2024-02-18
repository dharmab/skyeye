package audio

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/internal/debug"
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
	// TODO remove Oto debugging code
	otoCtx := debug.MustNewOtoContext()

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
				fb := debug.F32toBytes(txPCM)
				slog.Debug("playing audio", "S16LELen", len(fb))
				debug.PlayAudio(otoCtx, fb)
				slog.Info("publishing audio to receiver channel")
				c.rxchan <- txPCM
			}
		case <-ctx.Done():
			slog.Info("stopping voice decoder due to context cancellation")
			return
		}
	}
}

const (
	// frameLength is the length of an Opus frame sent by SRS.
	frameLength = 40 * time.Millisecond
	// sampleRate is the sample rate of the audio data sent by SRS in KHz
	sampleRate = 16000 // Wideband
	// channels is the number of channels in the audio data sent by SRS.
	channels = 1 // Mono
)

var frameSize = channels * frameLength.Milliseconds() * sampleRate / 1000 // 64KHz

// decode decodes the given Opus frame(s) into F32LE PCM audio data.
func (c *audioClient) decode(d *opus.Decoder, b []byte) ([]float32, error) {
	slog.Debug("decoding audio", "length", len(b))
	f32le := make([]float32, frameSize)
	n, err := d.DecodeFloat32(b, f32le)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Opus audio: %w", err)
	}
	f32le = f32le[:n*channels]
	slog.Debug("decoded samples", "count", n)
	return f32le, nil
}
