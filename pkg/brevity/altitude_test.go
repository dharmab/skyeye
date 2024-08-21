package brevity

import (
	"fmt"
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
			input:    []unit.Length{unit.Foot * 40},
			expected: []Stack{{Altitude: unit.Foot * 0, Count: 1}},
		},
		{
			input:    []unit.Length{unit.Foot * 100},
			expected: []Stack{{Altitude: unit.Foot * 100, Count: 1}},
		},
		{
			input:    []unit.Length{unit.Foot * 100, unit.Foot * 12000},
			expected: []Stack{{Altitude: unit.Foot * 12000, Count: 1}, {Altitude: unit.Foot * 100, Count: 1}},
		},
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
				assert.Equalf(t, expectedStack.Altitude, stack.Altitude, "expected %fft, got %fft", expectedStack.Altitude.Feet(), stack.Altitude.Feet())
				assert.Equal(t, expectedStack.Count, stack.Count, "stack count mismatch")
			}
		})
	}
}
