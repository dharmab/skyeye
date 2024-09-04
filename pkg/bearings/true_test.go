package bearings

import (
	"fmt"
	"math"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestNewTrueBearing(t *testing.T) {
	t.Parallel()
	for _, test := range testCompassBearings {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			a := unit.Angle(test.input) * unit.Degree
			bearing := NewTrueBearing(a)
			assert.InDelta(t, test.expectedDegrees, bearing.Value().Degrees(), 0.0001)
			assert.InDelta(t, test.expectedDegrees, bearing.Degrees(), 0.0001)
			assert.InDelta(t, bearing.Value().Degrees(), bearing.Degrees(), 0.0001)
			assert.InDelta(t, math.Round(test.expectedDegrees), bearing.Rounded().Degrees(), 0.0001)
			assert.InDelta(t, math.Round(test.expectedDegrees), bearing.RoundedDegrees(), 0.0001)
			assert.True(t, bearing.IsTrue())
			assert.False(t, bearing.IsMagnetic())
		})
	}
}

func TestTrueReciprocal(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    float64
		expected float64
	}{
		{-45, 135},
		{0, 180},
		{1, 181},
		{90, 270},
		{180, 360},
		{360, 180},
		{540, 360},
		{33.35, 213.35},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			a := unit.Angle(test.input) * unit.Degree
			bearing := NewTrueBearing(a)
			reciprocal := bearing.Reciprocal()
			assert.InDelta(t, test.expected, reciprocal.Degrees(), 0.0001)
		})
	}
}

func TestTrueMagnetic(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input       float64
		declination float64
		expected    float64
	}{
		{0, 0, 360},
		{360, 0, 360},
		{0, 1, 359},
		{2, 4, 358},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			a := unit.Angle(test.input) * unit.Degree
			d := unit.Angle(test.declination) * unit.Degree
			tru := NewTrueBearing(a)
			mag := tru.Magnetic(d)
			assert.InDelta(t, test.expected, mag.Degrees(), 0.0001)
		})
	}
}

func TestTrueString(t *testing.T) {
	t.Parallel()
	for _, test := range testCompassBearings {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			a := unit.Angle(test.input) * unit.Degree
			bearing := NewTrueBearing(a)
			assert.Equal(t, test.expectedString, bearing.String())
		})
	}
}
