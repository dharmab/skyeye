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

		// Homophones: "won" misheard for "one"
		{"Eagle won won", "eagle 1 1"}, //nolint:dupword
		{"Eagle won 1", "eagle 1 1"},
		{"Eagle 1 won", "eagle 1 1"},

		// Homophones: "to"/"too"/"tu" misheard for "two"
		{"Eagle to 1", "eagle 2 1"},
		{"Eagle too 1", "eagle 2 1"},
		{"Eagle 1 to", "eagle 1 2"},
		{"Eagle 1 too", "eagle 1 2"},
		{"Eagle to to", "eagle 2 2"},   //nolint:dupword
		{"Eagle too too", "eagle 2 2"}, //nolint:dupword

		// Homophones: "for"/"fore" misheard for "four"
		{"Eagle for 1", "eagle 4 1"},
		{"Eagle 1 for", "eagle 1 4"},
		{"Eagle for for", "eagle 4 4"}, //nolint:dupword
		{"Eagle fore 1", "eagle 4 1"},

		// Homophones: "free"/"tree" misheard for "three"
		{"Eagle free 1", "eagle 3 1"},
		{"Eagle 1 free", "eagle 1 3"},
		{"Eagle tree 1", "eagle 3 1"},

		// Homophones: "ate" misheard for "eight"
		{"Eagle ate 1", "eagle 8 1"},
		{"Eagle 1 ate", "eagle 1 8"},

		// Homophones: "niner" misheard for "nine"
		{"Eagle niner 1", "eagle 9 1"},

		// Ordinals misheard for digits
		{"Eagle 1st", "eagle 1"},
		{"Eagle 2nd", "eagle 2"},
		{"Eagle 3rd", "eagle 3"},
		{"Eagle 4th", "eagle 4"},
		{"Eagle 5th", "eagle 5"},
		{"Eagle 6th", "eagle 6"},
		{"Eagle 7th", "eagle 7"},
		{"Eagle 8th", "eagle 8"},
		{"Eagle 9th", "eagle 9"},

		// Mixed homophones: all digit combinations 1-9 x 1-9
		// using commonly misheard forms
		{"Eagle won to", "eagle 1 2"},
		{"Eagle won free", "eagle 1 3"},
		{"Eagle won for", "eagle 1 4"},
		{"Eagle to free", "eagle 2 3"},
		{"Eagle to for", "eagle 2 4"},
		{"Eagle free to", "eagle 3 2"},
		{"Eagle free for", "eagle 3 4"},
		{"Eagle for to", "eagle 4 2"},
		{"Eagle for free", "eagle 4 3"},
		{"Eagle for ate", "eagle 4 8"},
		{"Eagle ate to", "eagle 8 2"},
		{"Eagle ate for", "eagle 8 4"},
		{"Eagle ate ate", "eagle 8 8"}, //nolint:dupword

		// Deduplicate repeated callsign name from STT stuttering
		{"Eagle Eagle 2 7", "eagle 2 7"},    //nolint:dupword
		{"Eagle eagle 2 7", "eagle 2 7"},    //nolint:dupword
		{"Viper Viper 3 1", "viper 3 1"},    //nolint:dupword
		{"Falcon falcon 1 2", "falcon 1 2"}, //nolint:dupword
		{"Hornet Hornet 4 1", "hornet 4 1"}, //nolint:dupword
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
