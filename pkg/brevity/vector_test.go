package brevity

import (
	"strconv"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestVectorRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    unit.Length
		expected float64
	}{
		{input: 10 * unit.NauticalMile, expected: 10},
		{input: 10.4 * unit.NauticalMile, expected: 10},
		{input: 10.5 * unit.NauticalMile, expected: 11},
		{input: 0 * unit.NauticalMile, expected: 0},
		{input: 1 * unit.NauticalMile, expected: 1},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			v := NewVector(bearings.NewMagneticBearing(90*unit.Degree), test.input)
			assert.InDelta(t, test.expected, v.Range().NauticalMiles(), 0.001)
		})
	}
}

func TestVectorBearing(t *testing.T) {
	t.Parallel()
	bearing := bearings.NewMagneticBearing(270 * unit.Degree)
	v := NewVector(bearing, 50*unit.NauticalMile)
	assert.InDelta(t, 270.0, v.Bearing().Degrees(), 0.001)
	assert.True(t, v.Bearing().IsMagnetic())
}

func TestVectorRequestString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		request  VectorRequest
		expected string
	}{
		{
			request:  VectorRequest{Callsign: "eagle 1", Location: "home plate"},
			expected: "VECTOR to home plate for eagle 1",
		},
		{
			request:  VectorRequest{Callsign: "viper 1", Location: "tanker"},
			expected: "VECTOR to tanker for viper 1",
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, test.request.String())
		})
	}
}
