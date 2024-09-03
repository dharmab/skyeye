package coalitions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpposite(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    Coalition
		expected Coalition
	}{
		{Red, Blue},
		{Blue, Red},
		{Neutrals, Neutrals},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprint(test.input), func(t *testing.T) {
			t.Parallel()
			actual := test.input.Opposite()
			assert.Equal(t, test.expected, actual)
		})
	}
}
