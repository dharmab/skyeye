package speakers

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
	"github.com/go-audio/aiff"
	"github.com/martinlindhe/unit"
)

type macOSSynth struct {
	voice string
	rate  *unit.Frequency
}

var _ Speaker = (*macOSSynth)(nil)

func NewMacOSSpeaker(_ voices.Voice, playbackSpeed float64) Speaker {
	synth := &macOSSynth{
		voice: "Samantha",
	}

	if playbackSpeed != 1.0 {
		const (
			maxRate     = 300 * unit.Hertz
			defaultRate = 180 * unit.Hertz
			minRate     = 120 * unit.Hertz
		)
		var rate unit.Frequency
		if playbackSpeed < 0 {
			rate = maxRate
		} else if playbackSpeed > 1 {
			rate = minRate
		} else {
			var shift unit.Frequency
			if playbackSpeed < 1.0 {
				shift = unit.Frequency(playbackSpeed*(maxRate-defaultRate).Hertz()) * unit.Hertz
			} else {
				shift = unit.Frequency(1-playbackSpeed*(maxRate-defaultRate).Hertz()) * unit.Hertz
			}
			rate = defaultRate + shift
		}
		synth.rate = &rate
	}

	return synth
}

func (s *macOSSynth) Say(text string) ([]float32, error) {
	outFile, err := os.CreateTemp("", "skyeye-*.aiff")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary AIFF file: %w", err)
	}
	defer os.Remove(outFile.Name())

	args := []string{"--voice", s.voice, "--output", outFile.Name()}
	if s.rate != nil {
		args = append(args, "--rate", fmt.Sprintf("%.1f", s.rate.Hertz()))
	}
	args = append(args, text)
	command := exec.Command("say", args...)
	if err = command.Run(); err != nil {
		return nil, fmt.Errorf("failed to execute 'say' command: %w", err)
	}

	decoder := aiff.NewDecoder(outFile)
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode AIFF file: %w", err)
	}
	data := buf.AsFloat32Buffer().Data
	return data, nil
}
