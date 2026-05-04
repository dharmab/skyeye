package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{
			text: "Anyface, eagle 1, vector to tanker",
			expected: &brevity.VectorRequest{
				Callsign: "eagle 1",
				Location: "tanker",
			},
		},
		{
			text: "Anyface, eagle 1, vector tanker",
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

func TestParserVectorAmbiguityPrefersLongestSpan(t *testing.T) {
	t.Parallel()
	// Both "home" and "home plate" are configured. The longer match should
	// win regardless of iteration order — this is the guarantee that makes
	// matching deterministic.
	locations := []string{"home", "home plate"}
	parser := New(TestCallsign, locations, true)
	request := parser.Parse("Anyface, eagle 1, vector to home plate")
	require.IsType(t, &brevity.VectorRequest{}, request)
	assert.Equal(t, "home plate", request.(*brevity.VectorRequest).Location)
}

func TestParserVectorUnableCases(t *testing.T) {
	t.Parallel()
	locations := []string{"home plate"}
	parser := New(TestCallsign, locations, true)

	// "vector to" alone — no candidate location at all.
	res := parser.Parse("Anyface, eagle 1, vector to")
	require.IsType(t, &brevity.UnableToUnderstandRequest{}, res)

	// Location not configured.
	res = parser.Parse("Anyface, eagle 1, vector to atlantis")
	require.IsType(t, &brevity.UnableToUnderstandRequest{}, res)
}
