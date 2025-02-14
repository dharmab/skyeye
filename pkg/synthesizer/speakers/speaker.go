// Package speakers contains interfaces and implementations for text-to-speech speakers.
package speakers

import "github.com/martinlindhe/unit"

// Speaker provides text-to-speech.
type Speaker interface {
	// Say returns F32LE PCM audio for the given text.
	Say(string) ([]float32, error)
	// Close releases resources.
	Close() error
}

const targetSampleRate = 16000 * unit.Hertz
