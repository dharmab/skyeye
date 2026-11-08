//go:build !darwin || !cgo

package recognizer

import "errors"

// NewAppleSpeechRecognizer requires macOS with cgo and on-device English speech recognition.
func NewAppleSpeechRecognizer(_ string, _ ...Option) (Recognizer, error) {
	return nil, errors.New("Apple Speech requires macOS with cgo enabled")
}
