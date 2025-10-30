package piper

import (
	"context"
	"fmt"
	"time"

	asset "github.com/amitybell/piper-asset"
	masculine "github.com/amitybell/piper-voice-alan"
	feminine "github.com/amitybell/piper-voice-jenny"
	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/martinlindhe/unit"
	"github.com/nabbl/piper"
)

type speaker struct {
	tts           *piper.TTS
	speed         float64
	pauseDuration time.Duration
	gain          float64
}

var _ speakers.Speaker = (*speaker)(nil)

// Option is a functional option for configuring a Piper speaker.
type Option func(*speaker) error

// WithVoice configures the voice to use for the Piper speaker.
func WithVoice(v voices.Voice) Option {
	return func(s *speaker) error {
		var a asset.Asset
		if v == voices.MasculineVoice {
			a = masculine.Asset
		} else {
			a = feminine.Asset
		}
		tts, err := piper.New("", a)
		if err != nil {
			return fmt.Errorf("failed to create speaker: %w", err)
		}
		s.tts = tts
		return nil
	}
}

// WithSpeed configures the playback speed for the Piper speaker.
func WithSpeed(speed float64) Option {
	return func(s *speaker) error {
		s.speed = speed
		return nil
	}
}

// WithPause configures the pause duration for the Piper speaker.
func WithPause(duration time.Duration) Option {
	return func(s *speaker) error {
		s.pauseDuration = duration
		return nil
	}
}

// WithGain configures the gain (volume multiplier) for the Piper speaker.
func WithGain(gain float64) Option {
	return func(s *speaker) error {
		s.gain = gain
		return nil
	}
}

// New creates a Speaker powered by Piper (https://github.com/rhasspy/piper)
func New(opts ...Option) (speakers.Speaker, error) {
	s := &speaker{}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// SayContext implements [speakers.Speaker.SayContext].
func (s *speaker) SayContext(_ context.Context, text string) ([]float32, error) {
	synthesized, err := s.tts.Synthesize(text, piper.WithSpeed(float32(s.speed)), piper.WithPause(float32(s.pauseDuration.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize text: %w", err)
	}
	downsampled, err := speakers.Downsample(synthesized, 22050*unit.Hertz)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample synthesized audio: %w", err)
	}
	f32le := pcm.S16LEBytesToF32LE(downsampled)
	f32le = pcm.F32LEGain(f32le, s.gain)
	return f32le, nil
}

// Say implements [speakers.Speaker.Say].
func (s *speaker) Say(text string) ([]float32, error) {
	return s.SayContext(context.Background(), text)
}
