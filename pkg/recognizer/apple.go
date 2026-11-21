package recognizer

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
)

// appleContextualStrings supplies vocabulary rather than a Whisper-style instruction.
func appleContextualStrings(callsign string, opts ...Option) []string {
	var options recognizerOptions
	for _, opt := range opts {
		opt(&options)
	}
	result := []string{}
	seen := map[string]bool{}
	for _, phrase := range append([]string{callsign, "ANYFACE", "BULLSEYE", "BRAA", "BOGEY", "DECLARE", "SNAPLOCK", "SPIKED"}, options.locations...) {
		phrase = strings.TrimSpace(phrase)
		key := strings.ToLower(phrase)
		if phrase != "" && !seen[key] {
			result = append(result, phrase)
			seen[key] = true
		}
	}
	return result
}

func appleValidate(ctx context.Context, pcm []float32) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(pcm) == 0 {
		return errors.New("apple speech requires nonempty mono PCM audio")
	}
	if uint64(len(pcm)) > uint64(math.MaxUint32) {
		return errors.New("apple speech PCM exceeds native frame capacity")
	}
	for i, sample := range pcm {
		if math.IsNaN(float64(sample)) || math.IsInf(float64(sample), 0) || sample < -1 || sample > 1 {
			return fmt.Errorf("apple speech PCM sample %d must be finite and within [-1, 1]", i)
		}
	}
	return ctx.Err()
}
