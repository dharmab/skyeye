package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParserDeclare(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "anyface, tater 1-1, declare bullseye 0-5-4, 123, 3000",
			expected: &brevity.DeclareRequest{
				Callsign: "tater 1 1",
				Bullseye: *brevity.NewBullseye(
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
				Bullseye: *brevity.NewBullseye(
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
				Bullseye: *brevity.NewBullseye(
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
				Bullseye: *brevity.NewBullseye(
					bearings.NewMagneticBearing(75*unit.Degree),
					26*unit.NauticalMile,
				),
				Altitude: 2000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
		{
			text: "Anyface Goblin11, declare 052-77-2000",
			expected: &brevity.DeclareRequest{
				Callsign: "goblin 1 1",
				Bullseye: *brevity.NewBullseye(
					bearings.NewMagneticBearing(52*unit.Degree),
					77*unit.NauticalMile,
				),
				Altitude: 2000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
		},
		{
			text: "anyface, Chaos 11, declare braa 176 24 3000",
			expected: &brevity.DeclareRequest{
				Callsign: "chaos 1 1",
				Bearing:  bearings.NewMagneticBearing(176 * unit.Degree),
				Range:    24 * unit.NauticalMile,
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
				IsBRAA:   true,
			},
		},
		{
			text: "anyface, Chaos 11, declare braa 176 24 3000",
			expected: &brevity.DeclareRequest{
				Callsign: "chaos 1 1",
				Bearing:  bearings.NewMagneticBearing(176 * unit.Degree),
				Range:    24 * unit.NauticalMile,
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
				IsBRAA:   true,
			},
		},
		{
			text: "anyface, Chaos 11, declare a contact at braa 176 24 3000",
			expected: &brevity.DeclareRequest{
				Callsign: "chaos 1 1",
				Bearing:  bearings.NewMagneticBearing(176 * unit.Degree),
				Range:    24 * unit.NauticalMile,
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
				IsBRAA:   true,
			},
		},
		{
			text: "Anyface. Scorpio 21. Declare. Bra 068, 116, 15,000.",
			expected: &brevity.DeclareRequest{
				Callsign: "scorpio 2 1",
				Bearing:  bearings.NewMagneticBearing(68 * unit.Degree),
				Range:    116 * unit.NauticalMile,
				Altitude: 15000 * unit.Foot,
				Track:    brevity.UnknownDirection,
				IsBRAA:   true,
			},
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.DeclareRequest)
		actual := request.(*brevity.DeclareRequest)
		assert.Equal(t, expected.Callsign, actual.Callsign)
		if expected.IsBRAA {
			assert.True(t, actual.IsBRAA)
			require.NotNil(t, actual.Bearing)
			assert.InDelta(t, expected.Bearing.Degrees(), actual.Bearing.Degrees(), 0.5)
			require.NotNil(t, actual.Range)
			assert.InDelta(t, expected.Range.NauticalMiles(), actual.Range.NauticalMiles(), 0.5)
		} else {
			assert.False(t, actual.IsBRAA)
			require.NotNil(t, actual.Bullseye)
			assert.InDelta(t, expected.Bullseye.Bearing().Degrees(), actual.Bullseye.Bearing().Degrees(), 0.5)
			assert.InDelta(t, expected.Bullseye.Distance().NauticalMiles(), actual.Bullseye.Distance().NauticalMiles(), 1)
		}
		assert.InDelta(t, expected.Altitude.Feet(), actual.Altitude.Feet(), 50)
		assert.Equal(t, expected.Track, actual.Track)
	})
}
