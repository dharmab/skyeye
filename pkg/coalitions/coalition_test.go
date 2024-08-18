package coalitions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpposite(t *testing.T) {
	testCases := []struct {
		input    Coalition
		expected Coalition
	}{
		{Red, Blue},
		{Blue, Red},
		{Neutrals, Neutrals},
	}

	for _, test := range testCases {
		actual := test.input.Opposite()
		assert.Equal(t, test.expected, actual)
	}
}
