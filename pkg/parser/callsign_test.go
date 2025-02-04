package parser

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{"[CLAN] Wolf 1", "wolf 1"},
		{"Wolf 1 [CLAN]", "wolf 1"},
		{"[CLAN] Wolf 1 [1SG]", "wolf 1"},
		{"[Wolf 1", "wolf 1"},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual, ok := ParsePilotCallsign(test.name)
			assert.True(t, ok)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParsePilotCallsignInvalid(t *testing.T) {
	t.Parallel()
	testCases := []string{
		"",
		"[]",
		"[CLAN]",
	}

	for i, test := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			_, ok := ParsePilotCallsign(test)
			assert.False(t, ok)
		})
	}
}
