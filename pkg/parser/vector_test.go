package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/assert"
)

func TestParserVector(t *testing.T) {
	t.Parallel()
	locations := []string{"home plate", "rock"}
	testCases := []parserTestCase{
		{
			text: "Anyface, Eagle 1, vector to home plate",
			expected: &brevity.VectorRequest{
				Callsign: "eagle 1",
				Location: "home plate",
			},
		},
		{
			text: "Anyface, eagle 1, vector rock",
			expected: &brevity.VectorRequest{
				Callsign: "eagle 1",
				Location: "rock",
			},
		},
		{
			text: "Anyface, eagle 1, vector to nearest tanker",
			expected: &brevity.VectorRequest{
				Callsign: "eagle 1",
				Location: "tanker",
			},
		},
	}
	parser := New(TestCallsign, locations, true)
	runParserTestCases(t, parser, testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.VectorRequest)
		actual := request.(*brevity.VectorRequest)
		assert.Equal(t, expected.Callsign, actual.Callsign)
		assert.Equal(t, expected.Location, actual.Location)
	})
}
