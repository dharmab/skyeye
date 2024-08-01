// package speakers contains interfaces and implementations for text-to-speech speakers.
package speakers

// Speaker provides text-to-speech.
type Speaker interface {
	// Say returns F32LE PCM audio for the given text.
	Say(string) ([]float32, error)
}
