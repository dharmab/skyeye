package pocket

import (
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
	"github.com/stretchr/testify/require"
)

var modelPath = flag.String("model-path", filepath.Join("..", "..", "..", "models", model.DirName), "path to Pocket TTS model directory")

// phrases contains realistic Speech text as produced by the composer.
var phrases = []struct {
	name string
	text string
}{
	{"RadioCheck", "FALCON 2 1, 5 by 5."},
	{"BogeyDope", "VIPER 1 1, Group bra 0 9 0, 40, 25000, flanking, track south, hostile, 2 contacts, Flanker."},
	{"Picture", "MAGIC, 3 groups. Group bullseye 0 9 0, 40, 25000, track east, hostile, 4 contacts, heavy, Flanker. Group bullseye 1 8 0, 30, 20000, track north, hostile, 2 contacts, Fulcrum. Group bullseye 2 7 0, 55, 15000, track west, hostile, MiG-29."},
	{"Merged", "EAGLE 2 FLIGHT, merged."},
	{"ThreatCall", "VIPER 1 FLIGHT, Threat group bra 1 8 0, 30, 25000, flanking, track north, hostile, 2 contacts, Fulcrum."},
	{"Clean", "HORNET 4 1, clean."},
	{"SunriseCall", "All players, GCI MAGIC sunrise on 2 5 1 point 0"},
}

func BenchmarkPocketSpeaker(b *testing.B) {
	speaker, err := New(*modelPath)
	require.NoError(b, err)
	defer speaker.Close()

	for _, phrase := range phrases {
		b.Run(phrase.name, func(b *testing.B) {
			for b.Loop() {
				_, err := speaker.Say(b.Context(), phrase.text)
				require.NoError(b, err)
			}
		})
	}
}

func BenchmarkPocketSpeakerThreads(b *testing.B) {
	b.Skip("enable manually to compare thread counts")
	threadCounts := []int{1, 2, 4, 8}

	for _, threads := range threadCounts {
		speaker, err := New(*modelPath, WithThreads(threads))
		require.NoError(b, err)

		b.Run(fmt.Sprintf("Threads%d", threads), func(b *testing.B) {
			for _, phrase := range phrases {
				b.Run(phrase.name, func(b *testing.B) {
					for b.Loop() {
						_, err := speaker.Say(b.Context(), phrase.text)
						require.NoError(b, err)
					}
				})
			}
		})

		speaker.Close()
	}
}
