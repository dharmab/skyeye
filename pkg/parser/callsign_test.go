package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
