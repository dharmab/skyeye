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
			expected: &brevity.BogeyDopeRequest{
				Callsign: "eagle 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: "anyface intruder 11 bogey dope fighters",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.FixedWing,
			},
		},
		{
			text: "anyface intruder 11 bogey dope just helos",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.RotaryWing,
			},
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.BogeyDopeRequest)
		actual := request.(*brevity.BogeyDopeRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.Equal(t, expected.Filter, actual.Filter)
	})
}
