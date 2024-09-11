package recognizer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/rs/zerolog/log"
)

type whisperRecognizer struct {
	model    whisper.Model
	callsign string
}

var _ Recognizer = &whisperRecognizer{}

// NewWhisperRecognizer creates a new recognizer using OpenAI Whisper.
func NewWhisperRecognizer(model *whisper.Model, callsign string) Recognizer {
	return &whisperRecognizer{model: *model}
}

const maxSize = 256 * 1024

// Recognize implements [Recognizer.Recognize] using whisper.cpp.
func (r *whisperRecognizer) Recognize(ctx context.Context, sample []float32) (string, error) {
	if len(sample) > maxSize {
		log.Warn().Int("length", len(sample)).Int("maxLength", maxSize).Msg("clamping sample to maximum size")
		sample = sample[:maxSize]
	}

	wCtx, err := r.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("error creating whisper context: %w", err)
	}
	prompt := fmt.Sprintf("You receive commands in this template: {Either ANYFACE or %s} {PILOT CALLSIGN} {DIGITS} {'RADIO' or 'ALPHA' or 'BOGEY' or 'PICTURE' or 'DECLARE' or 'SNAPLOCK' or 'SPIKED'} {ARGUMENTS}. Parse numbers as digits. Separate numbers if there is silence between them. You may hear keywords in the arguments such as BULLSEYE or BRAA.", r.callsign)
	wCtx.SetInitialPrompt(prompt)

	if wCtx.IsMultilingual() {
		_ = wCtx.SetLanguage("en")
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
		select {
		case <-ctx.Done():
			log.Warn().Msg("returning early from speech recognition due to context cancellation")
			return textBuilder.String(), nil
		default:
			segment, err := wCtx.NextSegment()
			if errors.Is(err, io.EOF) {
				return textBuilder.String(), nil
			}
			if err != nil {
				return textBuilder.String(), fmt.Errorf("error processing segment: %w", err)
			}
			textBuilder.WriteString(segment.Text)
		}
	}
}
