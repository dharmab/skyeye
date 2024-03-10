package recognizer

type Recognizer interface {
	Recognize([]float32) (string, error)
}
