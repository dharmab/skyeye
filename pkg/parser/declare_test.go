package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

func TestParserDeclare(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, tater 1-1, declare bullseye 0-5-4, 123, 3000",
			expectedRequest: &brevity.DeclareRequest{
				Callsign: "tater 1 1",
				Location: *brevity.NewBullseye(
					unit.Angle(54)*unit.Degree,
					unit.Length(123)*unit.NauticalMile,
				),
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
			expectedOk: true,
		},
		{
			text: "anyface Fox 1 2 declare bullseye 043 102 12,000",
			expectedRequest: &brevity.DeclareRequest{
				Callsign: "fox 1 2",
				Location: *brevity.NewBullseye(
					unit.Angle(43)*unit.Degree,
					unit.Length(102)*unit.NauticalMile,
				),
				Altitude: 12000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
			expectedOk: true,
		},
		{
			text: "anyface, Chaos 11, declare bullseye 076 44 3000.",
			expectedRequest: &brevity.DeclareRequest{
				Callsign: "chaos 1 1",
				Location: *brevity.NewBullseye(
					unit.Angle(76)*unit.Degree,
					unit.Length(44)*unit.NauticalMile,
				),
				Altitude: 3000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
			expectedOk: true,
		},
		{
			text: "anyface, dog one one, declare, bullseye 075-26-2000",
			expectedRequest: &brevity.DeclareRequest{
				Callsign: "dog 1 1",
				Location: *brevity.NewBullseye(
					unit.Angle(75)*unit.Degree,
					unit.Length(26)*unit.NauticalMile,
				),
				Altitude: 2000 * unit.Foot,
				Track:    brevity.UnknownDirection,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases)
}
