// Package speakers contains interfaces and implementations for text-to-speech speakers.
package speakers

import "context"

// Speaker provides text-to-speech.
type Speaker interface {
	// Say returns F32LE PCM audio for the given text.
	Say(context.Context, string) ([]float32, error)
}
