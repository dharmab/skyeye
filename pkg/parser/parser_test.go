package parser

import (
	"fmt"
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestCallsign = "Skyeye"

type parserTestCase struct {
	text     string
	expected any
}

func TestParsePilotCallsign(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		expected string
	}{
		{"jeff", "jeff"},
		{"Jeff", "jeff"},
		{"Wardog 1 4", "wardog 1 4"},
		{"Wardog 14", "wardog 1 4"},
		{"Wardog 1-4", "wardog 1 4"},
		{"WARDOG 14", "wardog 1 4"},
		{"WARDOG14", "wardog 1 4"},
		{"Mobius 1", "mobius 1"},
		{"Red 243", "red 2 4 3"},
		{"Red 054", "red 0 5 4"},
		{"Gunfighter request", "gunfighter"},
		{"This is Red 7", "red 7"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, ok := ParsePilotCallsign(test.name)
			require.True(t, ok)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func runParserTestCases(
	t *testing.T,
	p Parser,
	testCases []parserTestCase,
	fn func(*testing.T, parserTestCase, any),
) {
	t.Helper()
	for _, test := range testCases {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			t.Helper()
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
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.AlphaCheckRequest)
		actual := request.(*brevity.AlphaCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

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
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.RadioCheckRequest)
		actual := request.(*brevity.RadioCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

func TestIsSimilar(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			assert.Equal(t, test.expected, IsSimilar(test.a, test.b))
		})
	}
}
