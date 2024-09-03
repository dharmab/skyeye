package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFrequency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input             string
		expectedFrequency RadioFrequency
		expectedOk        bool
	}{
		{"", RadioFrequency{}, false},
		{"0", RadioFrequency{}, false},
		{"-1", RadioFrequency{}, false},
		{"30FM", RadioFrequency{30 * unit.Megahertz, types.ModulationFM}, true},
		{"30.0FM", RadioFrequency{30 * unit.Megahertz, types.ModulationFM}, true},
		{"251.0", RadioFrequency{251 * unit.Megahertz, types.ModulationAM}, true},
		{"251.0AM", RadioFrequency{251 * unit.Megahertz, types.ModulationAM}, true},
		{"251.1AM", RadioFrequency{251.1 * unit.Megahertz, types.ModulationAM}, true},
		{"251.1 AM", RadioFrequency{251.1 * unit.Megahertz, types.ModulationAM}, true},
		{"eekum bokum", RadioFrequency{}, false},
		{"AM", RadioFrequency{}, false},
		{"FM", RadioFrequency{}, false},
		{"0AM", RadioFrequency{}, false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()
			frequency, err := ParseRadioFrequency(test.input)
			if !test.expectedOk {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.InDelta(
				t,
				test.expectedFrequency.Frequency.Megahertz(),
				frequency.Frequency.Megahertz(),
				0.005,
			)
			assert.Equal(t, test.expectedFrequency.Modulation, frequency.Modulation)
		})
	}
}
