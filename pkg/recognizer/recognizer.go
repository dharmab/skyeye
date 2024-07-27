package recognizer

// Recogniser recognizes text from speech
type Recognizer interface {
	// Recognize takes PCMF32LE audio data and returns any recognized text.
	Recognize([]float32) (string, error)
}
