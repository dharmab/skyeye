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
			require.NotNil(t, actual)
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
		{
			text: "anyface eagle 1",
			expected: &brevity.UnableToUnderstandRequest{
				Callsign: "eagle 1",
			},
		},
	}
	runParserTestCases(
		t,
		New(TestCallsign, true),
		testCases,
		func(t *testing.T, test parserTestCase, request any) {
			t.Helper()
			expected := test.expected.(*brevity.UnableToUnderstandRequest)
			actual := request.(*brevity.UnableToUnderstandRequest)
			require.Equal(t, expected.Callsign, actual.Callsign)
		},
	)
}
