package controller

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollateCallsigns(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		receivers []string
		everyone  []string
		expected  []string
	}{
		{
			receivers: []string{},
			everyone:  []string{},
			expected:  []string{},
		},
		{
			receivers: []string{"eagle 1 1"},
			everyone:  []string{"eagle 1 1", "viper 1 1"},
			expected:  []string{"eagle 1 1"},
		},
		{
			receivers: []string{"eagle 1 1", "viper 1 1"},
			everyone:  []string{"eagle 1 1", "viper 1 1"},
			expected:  []string{"eagle 1 1", "viper 1 1"},
		},
		{
			receivers: []string{"eagle 1 1", "eagle 1 2"},
			everyone:  []string{"eagle 1 1", "eagle 1 2"},
			expected:  []string{"eagle 1 flight"},
		},
		{
			receivers: []string{"eagle 1 1", "eagle 1 2"},
			everyone:  []string{"eagle 1 1", "eagle 1 2", "eagle 1 3"},
			expected:  []string{"eagle 1 1", "1 2"},
		},
		{
			receivers: []string{"eagle 1 1", "eagle 1 3"},
			everyone:  []string{"eagle 1 1", "eagle 1 2", "eagle 1 3"},
			expected:  []string{"eagle 1 1", "1 3"},
		},
		{
			receivers: []string{"eagle 1 1", "eagle 1 2", "eagle 1 3", "eagle 1 4"},
			everyone: []string{
				"eagle 1 1", "eagle 1 2", "eagle 1 3", "eagle 1 4",
				"eagle 2 1", "eagle 2 2", "eagle 2 3", "eagle 2 4",
			},
			expected: []string{"eagle 1 flight"},
		},
		{
			receivers: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 2 1", "eagle 2 2",
			},
			everyone: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 2 1", "eagle 2 2",
				"viper 1 1", "viper 2 2",
			},
			expected: []string{
				"eagle 1 flight",
				"eagle 2 flight",
			},
		},
		{
			receivers: []string{
				"eagle 1 1", "eagle 1 2",
				"blaze",
			},
			everyone: []string{
				"eagle 1 1", "eagle 1 2",
				"blaze",
			},
			expected: []string{
				"eagle 1 flight",
				"blaze",
			},
		},
		{
			receivers: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 1",
			},
			everyone: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 1",
			},
			// Unlikely case, weird result
			expected: []string{
				"eagle 1 flight",
				"eagle 1",
			},
		},
		{
			receivers: []string{
				"eagle 1 1", "eagle 1 2",
			},
			everyone: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 1",
			},
			// Unlikely case, weird result
			expected: []string{
				"eagle 1 flight",
			},
		},
		{
			receivers: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 2",
			},
			everyone: []string{
				"eagle 1 1", "eagle 1 2",
				"eagle 2",
			},
			expected: []string{
				"eagle 1 flight",
				"eagle 2",
			},
		},
		{
			receivers: []string{"eagle 2 2", "eagle 2 1"},
			everyone:  []string{"eagle 2 2", "eagle 2 1", "eagle 2 3"},
			expected:  []string{"eagle 2 1", "2 2"},
		},
		{
			receivers: []string{"jackal 2 2", "jackal 2 1"},
			everyone:  []string{"jackal 2 2", "jackal 2 1", "jackal 2 3"},
			expected:  []string{"jackal 2 1", "2 2"},
		},
		{
			receivers: []string{"jackal 2 2", "jackal 2 1"},
			everyone:  []string{"jackal 2 2", "jackal 2 3", "jackal 2 1"},
			expected:  []string{"jackal 2 1", "2 2"},
		},
		{
			receivers: []string{"bravo 1 1", "alpha 1 1"},
			everyone:  []string{"bravo 1 1", "alpha 1 1"},
			expected:  []string{"alpha 1 1", "bravo 1 1"},
		},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			actual := collateCallsigns(testCase.receivers, testCase.everyone)
			assert.ElementsMatchf(t, testCase.expected, actual, "got: %v, expected: %v, receivers: %v, everyone: %v", actual, testCase.expected, testCase.receivers, testCase.everyone)
		})
	}
}
