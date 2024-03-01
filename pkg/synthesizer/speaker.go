package synthesizer

// Sythesizer provides text-to-speech.
type Sythesizer interface {
	// Say returns F32LE PCM audio for the given text.
	Say(string) ([]float32, error)
}
