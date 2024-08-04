package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/require"
)

func TestParserDeclare(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, tater 1-1, declare bullseye 0-5-4, 123, 3000",
			expected: &brevity.DeclareRequest{
				Callsign: "tater 1 1",
				Location: *brevity.NewBullseye(
					bearings.NewMagneticBearing(54*unit.Degree),
					123*unit.NauticalMile,
				),
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
		{
			text: "anyface Fox 1 2 declare bullseye 043 102 12,000",
			expected: &brevity.DeclareRequest{
				Callsign: "fox 1 2",
				Location: *brevity.NewBullseye(
					bearings.NewMagneticBearing(43*unit.Degree),
					102*unit.NauticalMile,
				),
				Altitude: 12000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
		{
			text: "anyface, Chaos 11, declare bullseye 076 44 3000.",
			expected: &brevity.DeclareRequest{
				Callsign: "chaos 1 1",
				Location: *brevity.NewBullseye(
					bearings.NewMagneticBearing(76*unit.Degree),
					44*unit.NauticalMile,
				),
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
		{
			text: "anyface, dog one one, declare, bullseye 075-26-2000",
			expected: &brevity.DeclareRequest{
				Callsign: "dog 1 1",
				Location: *brevity.NewBullseye(
					bearings.NewMagneticBearing(75*unit.Degree),
					26*unit.NauticalMile,
				),
				Altitude: 2000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.DeclareRequest)
		actual := request.(*brevity.DeclareRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.InDelta(t, expected.Location.Bearing().Degrees(), actual.Location.Bearing().Degrees(), 0.5)
		require.InDelta(t, expected.Location.Distance().NauticalMiles(), actual.Location.Distance().NauticalMiles(), 1)
		require.InDelta(t, expected.Altitude.Feet(), actual.Altitude.Feet(), 50)
		require.Equal(t, expected.Track, actual.Track)
	})
}
