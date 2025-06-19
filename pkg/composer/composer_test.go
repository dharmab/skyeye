package composer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b, expected string
	}{
		{"", "", ""},
		{"", "a", "a"},
		{"a", "b", "a b"},
		{"a ", "b", "a b"},
		{"a", " b", "a b"},
		{"a, ", "b", "a, b"},
		{"a,", " b", "a, b"},
		{"a,", "b", "a, b"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			actual := join(test.a, test.b)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestAddSpacing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input, expected string
	}{
		{"", ""},
		{"a", " a"},
		{"1", " 1"},
		{" a", " a"},
		{" 1", " 1"},
		{"!", "!"},
		{" ", " "},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			actual := addSpacing(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestToLower(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"A", "a"},
		{"a", "a"},
		{"1", "1"},
		{"AA", "aA"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			actual := lowerFirst(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
