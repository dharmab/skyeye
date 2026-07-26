package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
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

// Regression: the format verb was transposed ("%f.3%s"), and the frequency is stored in hertz, so
// 251MHz AM rendered as "251000000.000000.3AM".
func TestRadioFrequencyString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		frequency RadioFrequency
		expected  string
	}{
		{RadioFrequency{30 * unit.Megahertz, types.ModulationFM}, "30.000FM"},
		{RadioFrequency{44.5 * unit.Megahertz, types.ModulationFM}, "44.500FM"},
		{RadioFrequency{87.975 * unit.Megahertz, types.ModulationFM}, "87.975FM"},
		{RadioFrequency{116 * unit.Megahertz, types.ModulationAM}, "116.000AM"},
		{RadioFrequency{133.5 * unit.Megahertz, types.ModulationAM}, "133.500AM"},
		{RadioFrequency{151.975 * unit.Megahertz, types.ModulationAM}, "151.975AM"},
		{RadioFrequency{225 * unit.Megahertz, types.ModulationAM}, "225.000AM"},
		{RadioFrequency{251 * unit.Megahertz, types.ModulationAM}, "251.000AM"},
		{RadioFrequency{251.075 * unit.Megahertz, types.ModulationAM}, "251.075AM"},
		{RadioFrequency{399.975 * unit.Megahertz, types.ModulationAM}, "399.975AM"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, test.frequency.String())
		})
	}
}

func TestRadioFrequencyIsSameFrequency(t *testing.T) {
	t.Parallel()
	uhf := RadioFrequency{251 * unit.Megahertz, types.ModulationAM}

	assert.True(t, uhf.IsSameFrequency(RadioFrequency{251 * unit.Megahertz, types.ModulationAM}))
	assert.False(t, uhf.IsSameFrequency(RadioFrequency{251 * unit.Megahertz, types.ModulationFM}))
	assert.False(t, uhf.IsSameFrequency(RadioFrequency{133 * unit.Megahertz, types.ModulationAM}))
}

func TestClientFrequencies(t *testing.T) {
	t.Parallel()
	c := &Client{
		clientInfo: types.ClientInfo{
			RadioInfo: types.RadioInfo{Radios: []types.Radio{testRadio, testOtherRadio}},
		},
	}

	frequencies := c.Frequencies()

	require.Len(t, frequencies, 2)
	assert.InDelta(t, 251.0, frequencies[0].Frequency.Megahertz(), 0.001)
	assert.Equal(t, types.Modulation(types.ModulationAM), frequencies[0].Modulation)
	assert.InDelta(t, 133.0, frequencies[1].Frequency.Megahertz(), 0.001)
}

// The on-frequency counts gate whether the bot broadcasts at all, and the human/bot split keeps it
// from talking to an audience of other bots.
func TestClientsOnFrequency(t *testing.T) {
	t.Parallel()
	c := newSyncTestClient()
	c.syncClients([]types.ClientInfo{
		testPeer("peer000000000000000001", "Eagle 1", coalitions.Blue, testRadio),
		testPeer("peer000000000000000002", "Eagle 2", coalitions.Blue, testRadio),
		testPeer("peer000000000000000003", "GCI Sniper [BOT]", coalitions.Blue, testRadio),
		testPeer("peer000000000000000004", "Eagle 3", coalitions.Blue, testOtherRadio),
		testPeer("peer000000000000000005", "Bandit 1", coalitions.Red, testRadio),
	})

	assert.Equal(t, 3, c.ClientsOnFrequency(), "off-frequency and opposing coalition peers excluded")
	assert.Equal(t, 2, c.HumansOnFrequency())
	assert.Equal(t, 1, c.BotsOnFrequency())

	assert.True(t, c.IsOnFrequency("Eagle 1"))
	assert.False(t, c.IsOnFrequency("Eagle 3"), "tuned to a different frequency")
	assert.False(t, c.IsOnFrequency("Bandit 1"), "opposing coalition is never tracked")
	assert.False(t, c.IsOnFrequency("Nobody"))
}

func TestClientsOnFrequencyWhenEmpty(t *testing.T) {
	t.Parallel()
	c := newSyncTestClient()

	assert.Equal(t, 0, c.ClientsOnFrequency())
	assert.Equal(t, 0, c.HumansOnFrequency())
	assert.Equal(t, 0, c.BotsOnFrequency())
	assert.False(t, c.IsOnFrequency("Eagle 1"))
}
