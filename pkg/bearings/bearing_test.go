package bearings

import (
	"fmt"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
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
			actual := Normalize(unit.Angle(test.input) * unit.Degree).Degrees()
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestBearingToString(t *testing.T) {
	testCases := []struct {
		bearing  Bearing
		expected string
	}{
		{
			bearing:  NewTrueBearing(unit.Angle(0) * unit.Degree),
			expected: "360",
		},
		{

			bearing:  NewTrueBearing(unit.Angle(1) * unit.Degree),
			expected: "001",
		},
		{
			bearing:  NewTrueBearing(unit.Angle(10) * unit.Degree),
			expected: "010",
		},
		{
			bearing:  NewTrueBearing(unit.Angle(100) * unit.Degree),
			expected: "100",
		},
		{
			bearing:  NewTrueBearing(unit.Angle(359) * unit.Degree),
			expected: "359",
		},
		{
			bearing:  NewTrueBearing(unit.Angle(360) * unit.Degree),
			expected: "360",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bearing.String())
		})
	}
}
