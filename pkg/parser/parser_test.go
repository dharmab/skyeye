package parser

import (
	"fmt"
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
	p Parser,
	testCases []parserTestCase,
	fn func(*testing.T, parserTestCase, any),
) {
	for _, test := range testCases {
		t.Run(test.text, func(t *testing.T) {
			actual := p.Parse(test.text)
			require.IsType(t, test.expected, actual)
			fn(t, test, actual)
		})
	}
}
func TestParserSadPaths(t *testing.T) {
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
		New(TestCallsign),
		testCases,
		func(*testing.T, parserTestCase, any) {},
	)
}

func TestParserAlphaCheck(t *testing.T) {
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
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.AlphaCheckRequest)
		actual := request.(*brevity.AlphaCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

func TestParserRadioCheck(t *testing.T) {
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
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.RadioCheckRequest)
		actual := request.(*brevity.RadioCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

func TestIsSimilar(t *testing.T) {
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"SkyEye", "Sky Eye", true},
		{"Bandar", "Bandog", true},
		{"Sky Eye", "Ghost Eye", false},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.a, test.b), func(t *testing.T) {
			require.Equal(t, test.expected, IsSimilar(test.a, test.b))
		})
	}
}
