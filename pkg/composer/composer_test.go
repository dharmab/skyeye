package composer

import (
	"strings"
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

func TestApplyToFirstCharacter(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		f        func(string) string
		expected string
	}{
		{"", strings.ToLower, ""},
		{"A", strings.ToLower, "a"},
		{"a", strings.ToLower, "a"},
		{"1", strings.ToLower, "1"},
		{"AA", strings.ToLower, "aA"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			actual := applyToFirstCharacter(test.input, strings.ToLower)
			assert.Equal(t, test.expected, actual)
		})
	}
}
