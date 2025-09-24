package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

func TestParserCheckIn(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "anyface Baron91 checking in.",
			expected: &brevity.CheckInRequest{
				Callsign: "baron 9 1",
			},
		},
		{
			text: "anyface, Mako, 1-1, check in.",
			expected: &brevity.CheckInRequest{
				Callsign: "mako 1 1",
			},
		},
		{
			text: "anyface, FANG21, CHICKEN",
			expected: &brevity.CheckInRequest{
				Callsign: "fang 2 1",
			},
		},
	}

	runParserTestCases(t, New(TestCallsign, []string{}, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.CheckInRequest)
		actual := request.(*brevity.CheckInRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}
