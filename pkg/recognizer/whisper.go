package recognizer

import (
	"fmt"
	"io"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/rs/zerolog/log"
)

type whisperRecognizer struct {
	model whisper.Model
}

var _ Recognizer = &whisperRecognizer{}

// NewWhisperRecognizer creates a new recognizer using OpenAI Whisper
func NewWhisperRecognizer(model whisper.Model) Recognizer {
	return &whisperRecognizer{model: model}
}

// Recognize implements [Recognizer.Recognize] using whisper.cpp
func (r *whisperRecognizer) Recognize(sample []float32) (string, error) {
	wCtx, err := r.model.NewContext()
	wCtx.SetInitialPrompt("You are a Ground Control Intercept (GCI) operator. You receive text in the format [ANYFACE / CALLSIGN] [CALLSIGN] [DIGITS] [RADIO, ALPHA, BOGEY, PICTURE, DECLARE, SNAPLOCK, or SPIKED] MESSAGE. Parse numbers as digits. Do not use punctuation.")
	if err != nil {
		return "", fmt.Errorf("error creating context: %w", err)
	}

	err = wCtx.Process(
		sample,
		func(segment whisper.Segment) {
			log.Debug().Str("text", segment.Text).Msg("processing segment")
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("error processing sample: %w", err)
	}

	var textBuilder strings.Builder
	for {
		segment, err := wCtx.NextSegment()
		if err == io.EOF {
			break
		}
		if err != nil {
			return textBuilder.String(), fmt.Errorf("error processing segment: %w", err)
		}

		textBuilder.WriteString(fmt.Sprintf("%s\n", segment.Text))
	}
	return textBuilder.String(), nil
}
