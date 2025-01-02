package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

const TestCallsign = "Skyeye"

type parserTestCase struct {
	text     string
	expected any
}

func runParserTestCases(
	t *testing.T,
	p *Parser,
	testCases []parserTestCase,
	fn func(*testing.T, parserTestCase, any),
) {
	t.Helper()
	for _, test := range testCases {
		t.Run(test.text, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			actual := p.Parse(test.text)
			require.IsType(t, test.expected, actual)
			fn(t, test, actual)
		})
	}
}

func TestParserSadPaths(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text:     "anyface",
			expected: &brevity.UnableToUnderstandRequest{},
		},
		{
			text:     "anyface radio check",
			expected: &brevity.UnableToUnderstandRequest{},
		},
	}
	runParserTestCases(
		t,
		New(TestCallsign, true),
		testCases,
		func(*testing.T, parserTestCase, any) {},
	)
}

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
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.AlphaCheckRequest)
		actual := request.(*brevity.AlphaCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

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
	}

	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.CheckInRequest)
		actual := request.(*brevity.CheckInRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}
