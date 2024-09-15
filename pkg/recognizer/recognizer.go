// package recognizer recognizes text from speech
package recognizer

import "context"

// Recognizer recognizes text from speech.
type Recognizer interface {
	// Recognize takes PCMF32LE audio data and returns any recognized text.
	Recognize(ctx context.Context, pcm []float32, enableTranscriptionLogging bool) (string, error)
}
