package speakers

import (
	"bytes"
	"fmt"
	"time"

	asset "github.com/amitybell/piper-asset"
	masculine "github.com/amitybell/piper-voice-alan"
	feminine "github.com/amitybell/piper-voice-jenny"
	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/nabbl/piper"
	"github.com/zaf/resample"
)

type piperSynth struct {
	tts           *piper.TTS
	playbackSpeed float32
	playbackPause time.Duration
}

var _ Speaker = (*piperSynth)(nil)

// NewPiperSpeaker creates a Speaker powered by Piper (https://github.com/rhasspy/piper)
func NewPiperSpeaker(v voices.Voice, playbackSpeed float32, playbackPause time.Duration) (Speaker, error) {
	var a asset.Asset
	if v == voices.MasculineVoice {
		a = masculine.Asset
	} else {
		a = feminine.Asset
	}
	tts, err := piper.New("", a)
	if err != nil {
		return nil, fmt.Errorf("failed to create speaker: %w", err)
	}
	return &piperSynth{tts: tts, playbackSpeed: playbackSpeed, playbackPause: playbackPause}, nil
}

// Say implements [Speaker.Say].
func (s *piperSynth) Say(text string) ([]float32, error) {
	synthesized, err := s.tts.Synthesize(text, piper.WithSpeed(s.playbackSpeed), piper.WithPause(float32(s.playbackPause.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize text: %w", err)
	}
	downsampled, err := downsample(synthesized, 24000, 16000, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample synthesized audio: %w", err)
	}
	f32le := pcm.S16LEBytesToF32LE(downsampled)
	return f32le, nil
}

func downsample(in []byte, orignalRate float64, newRate float64, channels int) ([]byte, error) {
	var buf bytes.Buffer
	resampler, err := resample.New(&buf, orignalRate, newRate, channels, resample.I16, resample.LowQ)
	if err != nil {
		return nil, fmt.Errorf("failed to create resampler: %w", err)
	}
	defer resampler.Close()

	_, err = resampler.Write(in)
	if err != nil {
		return nil, fmt.Errorf("failed to resample synthesized audio: %w", err)
	}
	return buf.Bytes(), nil
}
