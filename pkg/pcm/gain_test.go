package pcm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecibelsToGainToDecibelsRoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		db float64
	}{
		{-60.0},
		{-40.0},
		{-20.0},
		{-12.0},
		{-6.0},
		{-3.0},
		{-1.0},
		{0.0},
		{1.0},
		{3.0},
		{6.0},
		{12.0},
		{20.0},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%.1fdB", test.db), func(t *testing.T) {
			t.Parallel()
			gain := DecibelsToGain(test.db)
			result := GainToDecibels(gain)
			assert.InDelta(t, test.db, result, 1e-10, "got %v, expected %v, intermediate gain %v", result, test.db, gain)
		})
	}
}

func TestGainToDecibelsToGainRoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		gain float64
	}{
		{0.001},
		{0.01},
		{0.1},
		{0.25},
		{0.5},
		{0.707},
		{1.0},
		{1.414},
		{2.0},
		{4.0},
		{10.0},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%.3f", test.gain), func(t *testing.T) {
			t.Parallel()
			db := GainToDecibels(test.gain)
			result := DecibelsToGain(db)
			assert.InDelta(t, test.gain, result, 1e-10, "got %v, expected %v, intermediate dB %v", result, test.gain, db)
		})
	}
}
