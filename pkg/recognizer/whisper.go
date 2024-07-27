package recognizer

import (
	"fmt"
	"io"
	"strings"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
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
	ctx, err := r.model.NewContext()
	if err != nil {
		return "", fmt.Errorf("error creating context: %w", err)
	}

	err = ctx.Process(sample, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error processing sample: %w", err)
	}

	var textBuilder strings.Builder
	for {
		segment, err := ctx.NextSegment()
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
