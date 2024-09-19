package types

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/tacview/properties"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTimeFrame(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		line             string
		expectedDuration time.Duration
		expectedError    bool
	}{
		{
			line:             "#0.0000000",
			expectedDuration: 0,
		},
		{
			line:             "#122.510000",
			expectedDuration: 122*time.Second + 510*time.Millisecond,
		},
		{
			line:          "1b02,T=||-30.94",
			expectedError: true,
		},
		{
			line:          "-1c02",
			expectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.line, func(t *testing.T) {
			t.Parallel()
			actual, err := ParseTimeFrame(testCase.line)
			if testCase.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedDuration, actual)
			}
		})
	}
}

func TestParseObjectUpdate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		line           string
		expectedUpdate *ObjectUpdate
		expectedError  *error
	}{
		{
			line: "40000001,T=5.9005604|5.0219182|2000,Type=Navaid+Static+Bullseye,Color=Blue,Coalition=Enemies",
			expectedUpdate: &ObjectUpdate{
				ID: 0x40000001,
				Properties: map[string]string{
					properties.Transform: "5.9005604|5.0219182|2000",
					properties.Type:      "Navaid+Static+Bullseye",
					properties.Color:     "Blue",
					properties.Coalition: "Enemies",
				},
			},
		},
		{
			line: "40000002,T=5.9005604|5.0219182|2000,Type=Navaid+Static+Bullseye,Color=Grey,Coalition=Neutrals",
			expectedUpdate: &ObjectUpdate{
				ID: 0x40000002,
				Properties: map[string]string{
					properties.Transform: "5.9005604|5.0219182|2000",
					properties.Type:      "Navaid+Static+Bullseye",
					properties.Color:     "Grey",
					properties.Coalition: "Neutrals",
				},
			},
		},
		{
			line: "40000003,T=5.9004647|5.0217148|2000|-9.44|-22.3,Type=Navaid+Static+Bullseye,Color=Red,Coalition=Allies",
			expectedUpdate: &ObjectUpdate{
				ID: 0x40000003,
				Properties: map[string]string{
					properties.Transform: "5.9004647|5.0217148|2000|-9.44|-22.3",
					properties.Type:      "Navaid+Static+Bullseye",
					properties.Color:     "Red",
					properties.Coalition: "Allies",
				},
			},
		},
		{
			line: "14a02,T=5.5317022|3.2532355|598.99|||264.5|-40343.46|-195137.77|266.4,Type=Ground+Light+Human+Infantry,Name=Soldier M4,Pilot=Hound 1-1,Group=Hound 1,Color=Blue,Coalition=Enemies,Country=xb",
			expectedUpdate: &ObjectUpdate{
				ID: 0x14a02,
				Properties: map[string]string{
					properties.Transform: "5.5317022|3.2532355|598.99|||264.5|-40343.46|-195137.77|266.4",
					properties.Type:      "Ground+Light+Human+Infantry",
					properties.Name:      "Soldier M4",
					properties.Pilot:     "Hound 1-1",
					properties.Group:     "Hound 1",
					properties.Color:     "Blue",
					properties.Coalition: "Enemies",
					properties.Country:   "xb",
				},
			},
		},
		{
			line: "-14a02",
			expectedUpdate: &ObjectUpdate{
				ID:         0x14a02,
				IsRemoval:  true,
				Properties: map[string]string{},
			},
		},
		{
			line: "7a302,T=5.4407523|5.6618178|||3.6|84.5|-39416.41|72413.32|86.6",
			expectedUpdate: &ObjectUpdate{
				ID: 0x7a302,
				Properties: map[string]string{
					properties.Transform: "5.4407523|5.6618178|||3.6|84.5|-39416.41|72413.32|86.6",
				},
			},
		},
		{
			line: "1b02,T=||-30.94",
			expectedUpdate: &ObjectUpdate{
				ID: 0x1b02,
				Properties: map[string]string{
					properties.Transform: "||-30.94",
				},
			},
		},
		{
			line: "1c02,T=||-60.6",
			expectedUpdate: &ObjectUpdate{
				ID: 0x1c02,
				Properties: map[string]string{
					properties.Transform: "||-60.6",
				},
			},
		},
		{
			line: "a03,T=|||-34817.79|",
			expectedUpdate: &ObjectUpdate{
				ID: 0xa03,
				Properties: map[string]string{
					properties.Transform: "|||-34817.79|",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.line, func(t *testing.T) {
			t.Parallel()
			actual, err := ParseObjectUpdate(testCase.line)
			if testCase.expectedError != nil {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.expectedUpdate, actual)
			}
		})
	}
}
