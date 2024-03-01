package synthesizer

import (
	"bytes"
	"fmt"
	"log/slog"

	"github.com/amitybell/piper"
	asset "github.com/amitybell/piper-asset"
	masculine "github.com/amitybell/piper-voice-alan"
	feminine "github.com/amitybell/piper-voice-jenny"
	"github.com/dharmab/skyeye/pkg/pcm"
	"github.com/zaf/resample"
)

type piperSynth struct {
	tts *piper.TTS
}

var _ Sythesizer = (*piperSynth)(nil)

type Voice int

const (
	// FeminineVoice is the "Jenny" en-GB voice.
	// Origin: https://github.com/dioco-group/jenny-tts-dataset
	FeminineVoice Voice = iota
	// MasculineVoice is the "Alan" en-GB voice.
	// Origin: https://popey.me
	MasculineVoice
)

// NewPiperSpeaker creates a Speaker powered by Piper (https://github.com/rhasspy/piper)
func NewPiperSpeaker(v Voice) (Sythesizer, error) {
	var a asset.Asset
	if v == MasculineVoice {
		a = masculine.Asset
	} else {
		a = feminine.Asset
	}
	tts, err := piper.New("", a)
	if err != nil {
		return nil, fmt.Errorf("failed to create speaker: %w", err)
	}
	return &piperSynth{tts: tts}, nil
}

// Say implements Speaker.Say
func (s *piperSynth) Say(text string) ([]float32, error) {
	slog.Debug("synthesizing text", "text", text)
	synthesized, err := s.tts.Synthesize(text)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize text: %w", err)
	}
	slog.Debug("downsampling synthesized audio from 24KHz to 16KHz")
	downsampled, err := downsample(synthesized, 24000, 16000, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to downsample synthesized audio: %w", err)
	}
	slog.Debug("converting downsampled S16LE audio to F32LE")
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

	n, err := resampler.Write(in)
	if err != nil {
		return nil, fmt.Errorf("failed to resample synthesized audio: %w", err)
	}
	slog.Debug("resampled synthesized audio", "originalRate", orignalRate, "newRate", newRate, "length", n)
	return buf.Bytes(), nil
}
