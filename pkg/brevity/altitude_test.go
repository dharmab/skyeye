package brevity

import (
	"fmt"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestStacks(t *testing.T) {
	tests := []struct {
		input    []unit.Length
		expected []Stack
	}{
		{
			input:    []unit.Length{unit.Foot * 10000, unit.Foot * 20000, unit.Foot * 30000, unit.Foot * 40000, unit.Foot * 50000},
			expected: []Stack{{Altitude: unit.Foot * 50000, Count: 1}, {Altitude: unit.Foot * 40000, Count: 1}, {Altitude: unit.Foot * 30000, Count: 1}, {Altitude: unit.Foot * 20000, Count: 1}, {Altitude: unit.Foot * 10000, Count: 1}},
		},
		{
			input:    []unit.Length{unit.Foot * 10000, unit.Foot * 10000},
			expected: []Stack{{Altitude: unit.Foot * 10000, Count: 2}},
		},
		{
			input:    []unit.Length{unit.Foot * 10000, unit.Foot * 15000},
			expected: []Stack{{Altitude: unit.Foot * 15000, Count: 2}},
		},
		{
			input:    []unit.Length{unit.Foot * 10000, unit.Foot * 20000, unit.Foot * 20000},
			expected: []Stack{{Altitude: unit.Foot * 20000, Count: 2}, {Altitude: unit.Foot * 10000, Count: 1}},
		},
	}
	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			stacks := Stacks(test.input...)
			for i, stack := range stacks {
				expectedStack := test.expected[i]
				log.Trace().Float64("expected", float64(expectedStack.Altitude)).Float64("actual", float64(stack.Altitude)).Msg("checking altitudes")
				assert.Equal(t, expectedStack.Altitude, stack.Altitude)
				log.Trace().Int("expected", expectedStack.Count).Int("actual", stack.Count).Msg("checking counts")
				assert.Equal(t, expectedStack.Count, stack.Count)
			}
		})
	}
}
