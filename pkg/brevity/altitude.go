package brevity

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
)

// Stack represents a single layer of an altitude STACK.
type Stack struct {
	Altitude unit.Length
	Count    int
}

func (s Stack) String() string {
	return fmt.Sprintf("%d %.0f", s.Count, s.Altitude.Feet())
}

// Stacks creates altitude STACKS from altitudes.
func Stacks(altitudes ...unit.Length) []Stack {
	for i, alt := range altitudes {
		altitudes[i] = spatial.NormalizeAltitude(alt)
	}
	// reverse sort
	slices.SortFunc(altitudes, func(i, j unit.Length) int {
		if i < j {
			return -1
		}
		if i > j {
			return 1
		}
		return 0
	})

	stacks := []Stack{}
	for _, altitude := range slices.Backward(altitudes) {
		if altitude == 0 {
			continue
		}
		if len(stacks) == 0 {
			stacks = append(stacks, Stack{Altitude: altitude, Count: 1})
		} else {
			j := len(stacks) - 1
			highest := stacks[j].Altitude
			if altitude <= highest-9900*unit.Foot {
				stacks = append(stacks, Stack{Altitude: altitude, Count: 1})
			} else {
				stacks[j].Count++
			}
		}
	}

	return stacks
}
