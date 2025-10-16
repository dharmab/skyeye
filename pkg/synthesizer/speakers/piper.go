package speakers

import (
	"context"
	"fmt"
	"time"

	asset "github.com/amitybell/piper-asset"
	masculine "github.com/amitybell/piper-voice-alan"
	feminine "github.com/amitybell/piper-voice-jenny"
	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/martinlindhe/unit"
	"github.com/nabbl/piper"
)

type piperSynth struct {
	tts         *piper.TTS
	speed       float64
	pauseLength time.Duration
}

var _ Speaker = (*piperSynth)(nil)

// NewPiperSpeaker creates a Speaker powered by Piper (https://github.com/rhasspy/piper)
func NewPiperSpeaker(v voices.Voice, playbackSpeed float64, playbackPause time.Duration) (Speaker, error) {
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
	return &piperSynth{tts: tts, speed: playbackSpeed, pauseLength: playbackPause}, nil
}

// SayContext implements [Speaker.SayContext].
func (s *piperSynth) SayContext(_ context.Context, text string) ([]float32, error) {
	synthesized, err := s.tts.Synthesize(text, piper.WithSpeed(float32(s.speed)), piper.WithPause(float32(s.pauseLength.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize text: %w", err)
	}
	downsampled, err := downsample(synthesized, 22050*unit.Hertz)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample synthesized audio: %w", err)
	}
	f32le := pcm.S16LEBytesToF32LE(downsampled)
	return f32le, nil
}

// Say implements [Speaker.Say].
func (s *piperSynth) Say(text string) ([]float32, error) {
	return s.SayContext(context.Background(), text)
}
