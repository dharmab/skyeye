package bearings

import (
	"fmt"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    float64
		expected float64
	}{
		{
			input:    0,
			expected: 360,
		},
		{
			input:    1,
			expected: 1,
		},
		{
			input:    359,
			expected: 359,
		},
		{
			input:    360,
			expected: 360,
		},
		{
			input:    361,
			expected: 1,
		},
		{
			input:    -1,
			expected: 359,
		},
		{
			input:    360*4 + 90,
			expected: 90,
		},
		{
			input:    22.5,
			expected: 22.5,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			actual := normalize(unit.Angle(test.input) * unit.Degree).Degrees()
			require.InDelta(t, test.expected, actual, 0.1)
		})
	}
}

func TestBearingToString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		bearing  Bearing
		expected string
	}{
		{
			bearing:  NewTrueBearing(0 * unit.Degree),
			expected: "360",
		},
		{
			bearing:  NewTrueBearing(1 * unit.Degree),
			expected: "001",
		},
		{
			bearing:  NewTrueBearing(10 * unit.Degree),
			expected: "010",
		},
		{
			bearing:  NewTrueBearing(100 * unit.Degree),
			expected: "100",
		},
		{
			bearing:  NewTrueBearing(359 * unit.Degree),
			expected: "359",
		},
		{
			bearing:  NewTrueBearing(360 * unit.Degree),
			expected: "360",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.bearing.String())
		})
	}
}

var testCompassBearings = []struct {
	input           float64
	expectedDegrees float64
	expectedString  string
}{
	{-360, 360, "360"},
	{-90, 270, "270"},
	{-1, 359, "359"},
	{0, 360, "360"},
	{1, 1, "001"},
	{5, 5, "005"},
	{10, 10, "010"},
	{15, 15, "015"},
	{20, 20, "020"},
	{30, 30, "030"},
	{40, 40, "040"},
	{45, 45, "045"},
	{50, 50, "050"},
	{60, 60, "060"},
	{70, 70, "070"},
	{80, 80, "080"},
	{90, 90, "090"},
	{100, 100, "100"},
	{110, 110, "110"},
	{120, 120, "120"},
	{130, 130, "130"},
	{135, 135, "135"},
	{140, 140, "140"},
	{150, 150, "150"},
	{160, 160, "160"},
	{170, 170, "170"},
	{180, 180, "180"},
	{190, 190, "190"},
	{200, 200, "200"},
	{210, 210, "210"},
	{220, 220, "220"},
	{225, 225, "225"},
	{230, 230, "230"},
	{240, 240, "240"},
	{250, 250, "250"},
	{260, 260, "260"},
	{270, 270, "270"},
	{280, 280, "280"},
	{290, 290, "290"},
	{300, 300, "300"},
	{310, 310, "310"},
	{320, 320, "320"},
	{330, 330, "330"},
	{340, 340, "340"},
	{350, 350, "350"},
	{360, 360, "360"},
	{361, 1, "001"},
	{540, 180, "180"},
	{720, 360, "360"},
	{1080, 360, "360"},
	{34.5, 34.5, "035"},
	{33.49, 33.49, "033"},
}
