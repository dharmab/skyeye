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
		// Nothing more to do, since the type is already checked
		func(*testing.T, parserTestCase, any) {},
	)
}
