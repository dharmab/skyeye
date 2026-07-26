package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadioInfoIsOnFrequency(t *testing.T) {
	t.Parallel()
	var (
		uhf      = Radio{Frequency: 251_000_000, Modulation: ModulationAM}
		vhf      = Radio{Frequency: 133_000_000, Modulation: ModulationAM}
		fm       = Radio{Frequency: 30_000_000, Modulation: ModulationFM}
		guardVHF = Radio{Frequency: 121_500_000, Modulation: ModulationAM}
	)

	tests := []struct {
		name     string
		this     []Radio
		other    []Radio
		expected bool
	}{
		{
			name:     "single shared frequency",
			this:     []Radio{uhf},
			other:    []Radio{uhf},
			expected: true,
		},
		{
			name:     "matches on any radio in the inventory",
			this:     []Radio{uhf, vhf, fm},
			other:    []Radio{guardVHF, fm},
			expected: true,
		},
		{
			name:     "no shared frequency",
			this:     []Radio{uhf, vhf},
			other:    []Radio{fm, guardVHF},
			expected: false,
		},
		{
			// Observed on a live server: ATIS clients appear with no radios at all.
			name:     "other has no radios",
			this:     []Radio{uhf},
			other:    nil,
			expected: false,
		},
		{
			name:     "this has no radios",
			this:     nil,
			other:    []Radio{uhf},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			this := RadioInfo{Radios: test.this}
			other := RadioInfo{Radios: test.other}
			assert.Equal(t, test.expected, this.IsOnFrequency(other))
			assert.Equal(t, test.expected, other.IsOnFrequency(this), "must be symmetric")
		})
	}
}
