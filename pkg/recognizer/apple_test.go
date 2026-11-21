package recognizer

import (
	"context"
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestAppleValidation(t *testing.T) {
	t.Parallel()

	for _, pcm := range [][]float32{nil, {}, {float32(math.NaN())}, {float32(math.Inf(1))}, {1.01}, {-1.01}} {
		if err := appleValidate(context.Background(), pcm); err == nil {
			t.Errorf("accepted invalid PCM: %v", pcm)
		}
	}
	if err := appleValidate(context.Background(), []float32{-1, 0, 1}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := appleValidate(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}

func TestAppleContextualStrings(t *testing.T) {
	t.Parallel()

	locations := []string{"  Batumi ", "batumi", "", "MAGIC"}
	got := appleContextualStrings(" Magic ", WithLocations(locations))
	want := []string{"Magic", "ANYFACE", "BULLSEYE", "BRAA", "BOGEY", "DECLARE", "SNAPLOCK", "SPIKED", "Batumi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	locations[0] = "changed"
	if got[len(got)-1] != "Batumi" {
		t.Fatal("retained mutable options")
	}
}
