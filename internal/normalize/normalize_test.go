package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"Eagle 12 | Raptor", "eagle 12"},
		{"CASING", "casing"},
		{"foo-bar_baz!@#$%^&*()_+=", "foo bar baz"},
		{"a1b2c3", "a 1b 2c 3"},
		{"  Eagle 12  ", "eagle 12"},
		{"Eagle  12", "eagle 12"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			actual := Normalize(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSpaceDigits(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"Eagle11", "eagle 1 1"},
		{"305", "3 0 5"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			actual := SpaceDigits(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
