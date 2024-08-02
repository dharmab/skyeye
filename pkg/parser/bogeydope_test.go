package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

func TestParserBogeyDope(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 BOGEY DOPE",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "eagle 1",
				Filter:   brevity.Aircraft,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope fighters",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.FixedWing,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope just helos",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.RotaryWing,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expectedRequest.(*brevity.BogeyDopeRequest)
		actual := request.(*brevity.BogeyDopeRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.Equal(t, expected.Filter, actual.Filter)
	})
}
