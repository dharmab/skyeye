package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

func TestParserSnaplock(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, FREEDOM 31, SNAPLOCK 125 10, 8000",
			expectedRequest: &brevity.SnaplockRequest{
				Callsign: "freedom 3 1",
				BRA: brevity.NewBRA(
					125*unit.Degree,
					10*unit.NauticalMile,
					8000*unit.Foot,
				),
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases)
}
