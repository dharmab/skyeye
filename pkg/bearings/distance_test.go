package bearings

import (
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestAngularDistanceTrueBearings(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		a       float64
		b       float64
		wantDeg float64
	}{
		{"identical", 90, 90, 0},
		{"small delta", 10, 15, 5},
		{"wrap around due-north", 359, 1, 2},
		{"wrap around, swapped inputs", 1, 359, 2},
		{"opposite bearings", 0, 180, 180},
		{"large delta collapses via wraparound", 10, 350, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := NewTrueBearing(unit.Angle(tc.a) * unit.Degree)
			b := NewTrueBearing(unit.Angle(tc.b) * unit.Degree)
			got := AngularDistance(a, b).Degrees()
			assert.InDelta(t, tc.wantDeg, got, 0.01)
		})
	}
}

func TestAngularDistanceMagneticBearings(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		a       float64
		b       float64
		wantDeg float64
	}{
		{"identical", 90, 90, 0},
		{"small delta", 10, 15, 5},
		{"wrap around due-north", 359, 1, 2},
		{"wrap around, swapped inputs", 1, 359, 2},
		{"opposite bearings", 0, 180, 180},
		{"large delta collapses via wraparound", 10, 350, 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			a := NewMagneticBearing(unit.Angle(tc.a) * unit.Degree)
			b := NewMagneticBearing(unit.Angle(tc.b) * unit.Degree)
			got := AngularDistance(a, b).Degrees()
			assert.InDelta(t, tc.wantDeg, got, 0.01)
		})
	}
}
