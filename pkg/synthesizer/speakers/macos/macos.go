package macos

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	"github.com/go-audio/aiff"
	"github.com/martinlindhe/unit"
)

type speaker struct {
	rate  *unit.Frequency
	voice string
	gain  float64
}

var _ speakers.Speaker = (*speaker)(nil)

// Option is a functional option for configuring a macOS speaker.
type Option func(*speaker)

// WithSystemVoice configures the macOS speaker to use the system default voice.
// This can be customized by the user in System Preferences.
func WithSystemVoice() Option {
	return func(s *speaker) {
		s.voice = ""
	}
}

// WithSamanthaVoice configures the macOS speaker to use the "Samantha" voice,
// regardless of system settings.
func WithSamanthaVoice() Option {
	return func(s *speaker) {
		s.voice = "Samantha"
	}
}

// WithSpeed configures the playback speed for the macOS speaker.
// Values less than 1.0 increase speed, values greater than 1.0 decrease speed.
func WithSpeed(speed float64) Option {
	return func(s *speaker) {
		if speed != conf.DefaultPlaybackSpeed {
			const (
				maxRate     = 300 * unit.Hertz
				defaultRate = 180 * unit.Hertz
				minRate     = 120 * unit.Hertz
			)
			var rate unit.Frequency
			if speed < 0 {
				rate = maxRate
			} else if speed > conf.DefaultPlaybackSpeed {
				rate = minRate
			} else {
				var shift unit.Frequency
				if speed < conf.DefaultPlaybackSpeed {
					shift = unit.Frequency(speed*(maxRate-defaultRate).Hertz()) * unit.Hertz
				} else {
					shift = unit.Frequency(1-speed*(maxRate-defaultRate).Hertz()) * unit.Hertz
				}
				rate = defaultRate + shift
			}
			s.rate = &rate
		}
	}
}

// WithGain configures the gain (volume multiplier) for the macOS speaker.
func WithGain(gain float64) Option {
	return func(s *speaker) {
		s.gain = gain
	}
}

// New creates a Speaker powered by Apple's Speech Synthesis Manager.
func New(opts ...Option) speakers.Speaker {
	s := &speaker{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SayContext implements [speakers.Speaker.SayContext].
func (s *speaker) SayContext(ctx context.Context, text string) ([]float32, error) {
	outFile, err := os.CreateTemp("", "skyeye-*.aiff")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary AIFF file: %w", err)
	}
	defer os.Remove(outFile.Name())

	args := []string{"--output", outFile.Name()}
	if s.voice != "" {
		args = append(args, "--voice", s.voice)
	}
	if s.rate != nil {
		args = append(args, "--rate", fmt.Sprintf("%.1f", s.rate.Hertz()))
	}
	args = append(args, text)
	command := exec.CommandContext(ctx, "say", args...)
	if err = command.Run(); err != nil {
		return nil, fmt.Errorf("failed to execute 'say' command: %w", err)
	}

	decoder := aiff.NewDecoder(outFile)
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode AIFF file: %w", err)
	}
	f32Buf := buf.AsFloat32Buffer()
	f32Buf.Data = pcm.F32LEGain(f32Buf.Data, s.gain)
	b := pcm.F32LEToS16LEBytes(f32Buf.Data)
	sample, err := speakers.Downsample(b, unit.Frequency(decoder.SampleRate)*unit.Hertz)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample audio: %w", err)
	}

	f32le := pcm.S16LEBytesToF32LE(sample)
	return f32le, nil
}

// Say implements [speakers.Speaker.Say].
func (s *speaker) Say(text string) ([]float32, error) {
	return s.SayContext(context.Background(), text)
}
