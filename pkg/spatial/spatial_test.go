// spatial contains functions for working with the orb, bearings and unit modules together.

package spatial

import (
	"testing"

	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {
	testCases := []struct {
		a        orb.Point
		b        orb.Point
		expected unit.Length
	}{
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{0, 0},
			expected: 0,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{0, 1},
			expected: 111 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{0, -1},
			expected: 111 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{1, 0},
			expected: 111 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{-1, 0},
			expected: 111 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, 75},
			b:        orb.Point{1, 75},
			expected: 28.9 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, -75},
			b:        orb.Point{1, -75},
			expected: 28.9 * unit.Kilometer,
		},
		{
			a:        orb.Point{0, 90},
			b:        orb.Point{1, 90},
			expected: 0,
		},
		{
			a:        orb.Point{0, -90},
			b:        orb.Point{1, -90},
			expected: 0,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v -> %v", test.a, test.b), func(t *testing.T) {
			actual := Distance(test.a, test.b)
			assert.InDelta(t, test.expected.Kilometers(), actual.Kilometers(), 1)
		})
	}
}

func TestTrueBearing(t *testing.T) {
	testCases := []struct {
		a        orb.Point
		b        orb.Point
		expected unit.Angle
	}{
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{0, 1},
			expected: 360 * unit.Degree,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{1, 0},
			expected: 90 * unit.Degree,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{0, -1},
			expected: 180 * unit.Degree,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{-1, 0},
			expected: 270 * unit.Degree,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{1, 1},
			expected: 45 * unit.Degree,
		},
		{
			a:        orb.Point{0, 0},
			b:        orb.Point{-1, -1},
			expected: 225 * unit.Degree,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v -> %v", test.a, test.b), func(t *testing.T) {
			actual := TrueBearing(test.a, test.b)
			assert.InDelta(t, test.expected.Degrees(), actual.Degrees(), 1)
		})
	}
}

func TestIsZero(t *testing.T) {
	testCases := []struct {
		p        orb.Point
		expected bool
	}{
		{
			p:        orb.Point{0, 0},
			expected: true,
		},
		{
			p:        orb.Point{0, 1},
			expected: false,
		},
		{
			p:        orb.Point{1, 0},
			expected: false,
		},
		{
			p:        orb.Point{1, 1},
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v", test.p), func(t *testing.T) {
			actual := IsZero(test.p)
			assert.Equal(t, test.expected, actual)
		})
	}
}
func TestPointAtBearingAndDistance(t *testing.T) {
	testCases := []struct {
		origin   orb.Point
		bearing  bearings.Bearing
		distance unit.Length
		expected orb.Point
	}{
		{
			origin:   orb.Point{0, 0},
			bearing:  bearings.NewTrueBearing(90 * unit.Degree),
			distance: 0,
			expected: orb.Point{0, 0},
		},
		{
			origin:   orb.Point{0, 0},
			bearing:  bearings.NewTrueBearing(90 * unit.Degree),
			distance: 111 * unit.Kilometer,
			expected: orb.Point{1, 0},
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v, %v, %v", test.origin, test.bearing, test.distance), func(t *testing.T) {
			actual := PointAtBearingAndDistance(test.origin, test.bearing, test.distance)
			assert.InDelta(t, test.expected.Lon(), actual.Lon(), 0.01)
			assert.InDelta(t, test.expected.Lat(), actual.Lat(), 0.01)
		})
	}
}
