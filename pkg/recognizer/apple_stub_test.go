//go:build !darwin || !cgo

package recognizer

import "testing"

func TestAppleUnsupportedPlatform(t *testing.T) {
	t.Parallel()

	r, err := NewAppleSpeechRecognizer("Magic")
	if err == nil || r != nil {
		t.Fatalf("got (%v, %v), want unsupported-platform error", r, err)
	}
}
