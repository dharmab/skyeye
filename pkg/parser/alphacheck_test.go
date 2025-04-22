package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

func TestParserAlphaCheck(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "ANYFACE, HORNET 1, CHECKING IN AS FRAGGED, REQUEST ALPHA CHECK DEPOT",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "hornet 1",
			},
		},
		{
			text: "anyface intruder 11 alpha check",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface intruder 11, checking in as fragged, request alpha check bullseye",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: TestCallsign + "intruder 11 alpha check",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: TestCallsign + "Gunfighter 2-1, AlphaJack, Bullseye.",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "gunfighter 2 1",
			},
		},
		{
			text: TestCallsign + " HORNET 12, ALPHACHEK",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "hornet 1 2",
			},
		},
		{
			text: TestCallsign + ", Gunmetal 2-1, AlphaJack, Bullseye.",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "gunmetal 2 1",
			},
		},
		{
			text: TestCallsign + " Eagle 11, AlphaJuck.",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "eagle 1 1",
			},
		},
		{
			text: TestCallsign + " Eagle11, ARFA CHECK",
			expected: &brevity.AlphaCheckRequest{
				Callsign: "eagle 1 1",
			},
		},
	}
	runParserTestCases(t, New(TestCallsign, []string{}, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.AlphaCheckRequest)
		actual := request.(*brevity.AlphaCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}
