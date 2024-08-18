package bearings

import (
	"math"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestNewTrueBearing(t *testing.T) {
	testCases := []struct {
		input    float64
		expected float64
	}{
		{-360, 360},
		{-90, 270},
		{-1, 359},
		{0, 360},
		{1, 1},
		{5, 5},
		{10, 10},
		{15, 15},
		{20, 20},
		{30, 30},
		{40, 40},
		{45, 45},
		{50, 50},
		{60, 60},
		{70, 70},
		{80, 80},
		{90, 90},
		{100, 100},
		{110, 110},
		{120, 120},
		{130, 130},
		{135, 135},
		{140, 140},
		{150, 150},
		{160, 160},
		{170, 170},
		{180, 180},
		{190, 190},
		{200, 200},
		{210, 210},
		{220, 220},
		{225, 225},
		{230, 230},
		{240, 240},
		{250, 250},
		{260, 260},
		{270, 270},
		{280, 280},
		{290, 290},
		{300, 300},
		{310, 310},
		{320, 320},
		{330, 330},
		{340, 340},
		{350, 350},
		{360, 360},
		{361, 1},
		{540, 180},
		{720, 360},
		{1080, 360},
		{34.5, 34.5},
		{33.49, 33.49},
	}

	for _, test := range testCases {
		a := unit.Angle(test.input) * unit.Degree
		bearing := NewTrueBearing(a)
		assert.InDelta(t, test.expected, bearing.Value().Degrees(), 0.0001)
		assert.InDelta(t, test.expected, bearing.Degrees(), 0.0001)
		assert.InDelta(t, bearing.Value().Degrees(), bearing.Degrees(), 0.0001)
		assert.InDelta(t, math.Round(test.expected), bearing.Rounded().Degrees(), 0.0001)
		assert.InDelta(t, math.Round(test.expected), bearing.RoundedDegrees(), 0.0001)
		assert.True(t, bearing.IsTrue())
		assert.False(t, bearing.IsMagnetic())
	}
}

func TestTrueReciprocal(t *testing.T) {
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
		a := unit.Angle(test.input) * unit.Degree
		bearing := NewTrueBearing(a)
		reciprocal := bearing.Reciprocal()
		assert.InDelta(t, test.expected, reciprocal.Degrees(), 0.0001)
	}
}

func TestTrueMagnetic(t *testing.T) {
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
		a := unit.Angle(test.input) * unit.Degree
		d := unit.Angle(test.declination) * unit.Degree
		tru := NewTrueBearing(a)
		mag := tru.Magnetic(d)
		assert.InDelta(t, test.expected, mag.Degrees(), 0.0001)
	}
}

func TestTrueString(t *testing.T) {
	testCases := []struct {
		input    float64
		expected string
	}{
		{0, "360"},
		{1, "001"},
		{5, "005"},
		{10, "010"},
		{15, "015"},
		{20, "020"},
		{30, "030"},
		{40, "040"},
		{45, "045"},
		{50, "050"},
		{60, "060"},
		{70, "070"},
		{80, "080"},
		{90, "090"},
		{100, "100"},
		{110, "110"},
		{120, "120"},
		{130, "130"},
		{135, "135"},
		{140, "140"},
		{150, "150"},

		{160, "160"},
		{170, "170"},
		{180, "180"},
		{190, "190"},
		{200, "200"},
		{210, "210"},
		{220, "220"},
		{225, "225"},
		{230, "230"},
		{240, "240"},
		{250, "250"},
		{260, "260"},
		{270, "270"},
		{280, "280"},
		{290, "290"},
		{300, "300"},
		{310, "310"},
		{320, "320"},
		{330, "330"},
		{340, "340"},
		{350, "350"},
		{360, "360"},
		{361, "001"},
		{540, "180"},
		{720, "360"},
		{1080, "360"},
		{34.5, "035"},
		{33.49, "033"},
	}

	for _, test := range testCases {
		a := unit.Angle(test.input) * unit.Degree
		bearing := NewTrueBearing(a)
		assert.Equal(t, test.expected, bearing.String())
	}
}
