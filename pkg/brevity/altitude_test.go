package brevity

import (
	"strconv"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestStacks(t *testing.T) {
	tests := []struct {
		input    []unit.Length
		expected []Stack
	}{
		{
			input:    []unit.Length{},
			expected: []Stack{},
		},
		{
			input:    []unit.Length{0 * unit.Foot},
			expected: []Stack{},
		},
		{
			input:    []unit.Length{0 * unit.Foot, 100 * unit.Foot},
			expected: []Stack{{Altitude: 100 * unit.Foot, Count: 1}},
		},
		{
			input:    []unit.Length{40 * unit.Foot},
			expected: []Stack{},
		},
		{
			input:    []unit.Length{100 * unit.Foot},
			expected: []Stack{{Altitude: 100 * unit.Foot, Count: 1}},
		},
		{
			input:    []unit.Length{100 * unit.Foot, 12000 * unit.Foot},
			expected: []Stack{{Altitude: 12000 * unit.Foot, Count: 1}, {Altitude: 100 * unit.Foot, Count: 1}},
		},
		{
			input:    []unit.Length{10000 * unit.Foot, 20000 * unit.Foot, 30000 * unit.Foot, 40000 * unit.Foot, 50000 * unit.Foot},
			expected: []Stack{{Altitude: 50000 * unit.Foot, Count: 1}, {Altitude: 40000 * unit.Foot, Count: 1}, {Altitude: 30000 * unit.Foot, Count: 1}, {Altitude: 20000 * unit.Foot, Count: 1}, {Altitude: 10000 * unit.Foot, Count: 1}},
		},
		{
			input:    []unit.Length{10000 * unit.Foot, 10000 * unit.Foot},
			expected: []Stack{{Altitude: 10000 * unit.Foot, Count: 2}},
		},
		{
			input:    []unit.Length{10000 * unit.Foot, 15000 * unit.Foot},
			expected: []Stack{{Altitude: 15000 * unit.Foot, Count: 2}},
		},
		{
			input:    []unit.Length{10000 * unit.Foot, 20000 * unit.Foot, 20000 * unit.Foot},
			expected: []Stack{{Altitude: 20000 * unit.Foot, Count: 2}, {Altitude: 10000 * unit.Foot, Count: 1}},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			stacks := Stacks(test.input...)
			for i, stack := range stacks {
				expectedStack := test.expected[i]
				assert.InDelta(t, expectedStack.Altitude.Feet(), stack.Altitude.Feet(), 0.5)
				assert.Equal(t, expectedStack.Count, stack.Count, "stack count mismatch")
			}
		})
	}
}
