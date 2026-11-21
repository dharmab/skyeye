//go:build darwin && cgo

package recognizer

import (
	"context"
	"errors"
	"testing"
)

// Deliberately bypass the constructor and use only paths that cannot request
// authorization, even on a developer machine with an embedded usage description.
func TestAppleRecognizeWithoutAuthorization(t *testing.T) {
	t.Parallel()

	r := &appleSpeechRecognizer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Recognize(ctx, []float32{0}, false); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
	if _, err := r.Recognize(context.Background(), nil, false); err == nil {
		t.Fatal("accepted empty PCM")
	}
}
