package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSimilar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a        string
		b        string
		expected bool
	}{
		{"SkyEye", "Sky Eye", true},
		{"Bandar", "Bandog", true},
		{"Sky Eye", "Ghost Eye", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.a, test.b), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, isSimilar(test.a, test.b))
		})
	}
}
