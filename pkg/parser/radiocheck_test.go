package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

func TestParserRadioCheck(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "anyface Wildcat11 radio check out.",
			expected: &brevity.RadioCheckRequest{
				Callsign: "wildcat 1 1",
			},
		},
		{
			text: "Any face, Wildcat11, radio check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "wildcat 1 1",
			},
		},
		{
			text: "anyface intruder 11 radio check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface intruder 1-1 radio check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface intruder five one radio check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 5 1",
			},
		},
		{
			text: "anyface intruder 11 request radio check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface intruder 11 radio check 133 point zero",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface intruder 11 radio check on button five",
			expected: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: TestCallsign + ", this is Heaven11. Request Mic Check.",
			expected: &brevity.RadioCheckRequest{
				Callsign: "heaven 1 1",
			},
		},
		{
			text: TestCallsign + ", this is heaven. Request mic check",
			expected: &brevity.RadioCheckRequest{
				Callsign: "heaven",
			},
		},
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.RadioCheckRequest)
		actual := request.(*brevity.RadioCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}
