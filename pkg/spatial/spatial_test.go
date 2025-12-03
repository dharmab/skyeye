// spatial contains functions for working with the orb, bearings and unit modules together.
package spatial

import (
	"fmt"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestDistance(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a        orb.Point
		b        orb.Point
		expected unit.Length
	}{ // kola tests
		{
			a:        orb.Point{33.405794, 69.047461},
			b:        orb.Point{24.973478, 70.068836},
			expected: 186 * unit.NauticalMile,
		},
		
			{
				a:        orb.Point{33.405794, 69.047461},
				b:        orb.Point{34.262989, 64.91865},
				expected: 249 * unit.NauticalMile,
			},
			
			{
				a:        orb.Point{34.262989, 64.91865},
				b:        orb.Point{24.973478, 70.068836},
				expected: 377  * unit.NauticalMile,
			},
		/*
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
		*/
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v -> %v", test.a, test.b), func(t *testing.T) {
			t.Parallel()
			actual := Distance(test.a, test.b)
			assert.InDelta(t, test.expected.NauticalMiles(), actual.NauticalMiles(), 5)
		})
	}
}

func TestTrueBearing(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a        orb.Point
		b        orb.Point
		expected unit.Angle
	}{ // kola
		{
			a:        orb.Point{33.405794, 69.047461},
			b:        orb.Point{24.973478, 70.068836},
			expected: 282  * unit.Degree,
		},
		
			{
				a:        orb.Point{33.405794, 69.047461},
				b:        orb.Point{34.262989, 64.91865},
				expected: 164  * unit.Degree,
			},
			
			{
				a:        orb.Point{34.262989, 64.91865},
				b:        orb.Point{24.973478, 70.068836},
				expected: 317   * unit.Degree,
			},
			/*
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
			{
				//a:        orb.Point{69.047471, 33.405794},
				//b:        orb.Point{69.157219, 32.14515},
				a:        orb.Point{33.405794, 69.047471},
				b:        orb.Point{32.14515, 69.157219},
				expected: 274 * unit.Degree,
			},
		*/
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v -> %v", test.a, test.b), func(t *testing.T) {
			t.Parallel()
			actual := TrueBearing(test.a, test.b)
			assert.InDelta(t, test.expected.Degrees(), actual.Degrees(), 2)
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			actual := IsZero(test.p)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestPointAtBearingAndDistance(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			actual := PointAtBearingAndDistance(test.origin, test.bearing, test.distance)
			assert.InDelta(t, test.expected.Lon(), actual.Lon(), 0.01)
			assert.InDelta(t, test.expected.Lat(), actual.Lat(), 0.01)
		})
	}
}

func TestNormalizeAltitude(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    unit.Length
		expected unit.Length
	}{
		{
			input:    -100 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    0,
			expected: 0,
		},
		{
			input:    40 * unit.Foot,
			expected: 0,
		},
		{
			input:    100 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    120 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    200 * unit.Foot,
			expected: 200 * unit.Foot,
		},
		{
			input:    249 * unit.Foot,
			expected: 200 * unit.Foot,
		},
		{
			input:    250 * unit.Foot,
			expected: 300 * unit.Foot,
		},
		{
			input:    1234 * unit.Foot,
			expected: 1000 * unit.Foot,
		},
		{
			input:    10000 * unit.Foot,
			expected: 10000 * unit.Foot,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%fft", test.input.Feet()), func(t *testing.T) {
			t.Parallel()
			actual := NormalizeAltitude(test.input)
			assert.InDelta(t, test.expected.Feet(), actual.Feet(), 0.1)
		})
	}
}
