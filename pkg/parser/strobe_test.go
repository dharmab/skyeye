package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserStrobe(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 STROBE 2-7-0",
			expected: &brevity.StrobeRequest{
				Callsign: "eagle 1",
				Bearing:  bearings.NewMagneticBearing(unit.Angle(270) * unit.Degree),
			},
		},
		{
			text: "Anyface Raven 1-4, Strobe 0-2-0",
			expected: &brevity.StrobeRequest{
				Callsign: "raven 1 4",
				Bearing:  bearings.NewMagneticBearing(unit.Angle(20) * unit.Degree),
			},
		},
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.StrobeRequest)
		require.IsType(t, &brevity.StrobeRequest{}, request)
		actual := request.(*brevity.StrobeRequest)
		assert.Equal(t, expected.Callsign, actual.Callsign)
		assert.Equal(t, expected.Bearing, actual.Bearing)
	})
}
