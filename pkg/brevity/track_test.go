package brevity

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestTrackFromBearing(t *testing.T) {
	testCases := []struct {
		input    bearings.Bearing
		expected Track
	}{
		{bearings.NewMagneticBearing(0 * unit.Degree), North},
		{bearings.NewMagneticBearing(45 * unit.Degree), Northeast},
		{bearings.NewMagneticBearing(90 * unit.Degree), East},
		{bearings.NewMagneticBearing(135 * unit.Degree), Southeast},
		{bearings.NewMagneticBearing(180 * unit.Degree), South},
		{bearings.NewMagneticBearing(225 * unit.Degree), Southwest},
		{bearings.NewMagneticBearing(270 * unit.Degree), West},
		{bearings.NewMagneticBearing(315 * unit.Degree), Northwest},
		{bearings.NewMagneticBearing(360 * unit.Degree), North},
	}
	for _, test := range testCases {
		t.Run(test.input.String(), func(t *testing.T) {
			actual := TrackFromBearing(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
