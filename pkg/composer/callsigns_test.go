package composer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComposer_ComposeCallsigns(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		callsigns []string
		expected  []string
	}{
		{
			callsigns: []string{"Alpha 1"},
			expected:  []string{"ALPHA 1"},
		},
		{
			callsigns: []string{"Alpha 1", "Bravo 1", "Charlie 1"},
			expected:  []string{"ALPHA 1, BRAVO 1, CHARLIE 1"},
		},
		{
			callsigns: []string{"Alpha 2", "Alpha 1", "Alpha 3"},
			expected:  []string{"ALPHA 1, ALPHA 2, ALPHA 3"},
		},
		{
			callsigns: []string{"Charlie 1 1", "Alpha 1 2", "Bravo 1 2", "Bravo 1 1", "Alpha 1 1"},
			expected:  []string{"ALPHA 1 1, ALPHA 1 2", "BRAVO 1 1, BRAVO 1 2", "CHARLIE 1 1"},
		},
	}
	c := &composer{"Tester"}
	for _, testCase := range testCases {
		t.Run(strings.Join(testCase.callsigns, ", "), func(t *testing.T) {
			t.Parallel()
			for _, expected := range testCase.expected {
				assert.Contains(t, c.ComposeCallsigns(testCase.callsigns...), expected)
			}
		})
	}
}
