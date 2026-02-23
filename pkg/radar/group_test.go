package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

// makeTrackfile creates a trackfile with two frames so that Course() returns approximately
// the given true heading (degrees). Small magnetic declination may shift the magnetic
// course slightly, but cardinal directions remain stable across test locations.
func makeTrackfile(id uint64, trueBearingDegrees float64) *trackfiles.Trackfile {
	origin := orb.Point{-115.0, 36.0}
	bearing := bearings.NewTrueBearing(unit.Angle(trueBearingDegrees) * unit.Degree)
	dest := spatial.PointAtBearingAndDistance(origin, bearing, 1000*unit.Meter)
	now := time.Now()
	tf := trackfiles.New(trackfiles.Labels{
		ID:        id,
		ACMIName:  "F-15C",
		Name:      "Eagle",
		Coalition: coalitions.Blue,
	})
	tf.Update(trackfiles.Frame{
		Time:     now.Add(-1 * time.Second),
		Point:    origin,
		Altitude: 20000 * unit.Foot,
	})
	tf.Update(trackfiles.Frame{
		Time:     now,
		Point:    dest,
		Altitude: 20000 * unit.Foot,
	})
	return tf
}

func TestGroupTrack(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		group    group
		expected brevity.Track
	}{
		{
			name: "single contact heading north",
			group: group{
				contacts: []*trackfiles.Trackfile{
					makeTrackfile(1, 0),
				},
			},
			expected: brevity.North,
		},
		{
			name: "coherent crossing north (010 and 350)",
			group: group{
				contacts: []*trackfiles.Trackfile{
					makeTrackfile(1, 10),
					makeTrackfile(2, 350),
				},
			},
			expected: brevity.North,
		},
		{
			name: "incoherent opposites (090 and 270)",
			group: group{
				contacts: []*trackfiles.Trackfile{
					makeTrackfile(1, 90),
					makeTrackfile(2, 270),
				},
			},
			expected: brevity.UnknownDirection,
		},
		{
			name: "four evenly spread (N E S W)",
			group: group{
				contacts: []*trackfiles.Trackfile{
					makeTrackfile(1, 0),
					makeTrackfile(2, 90),
					makeTrackfile(3, 180),
					makeTrackfile(4, 270),
				},
			},
			expected: brevity.UnknownDirection,
		},
		{
			name: "two north two south",
			group: group{
				contacts: []*trackfiles.Trackfile{
					makeTrackfile(1, 0),
					makeTrackfile(2, 0),
					makeTrackfile(3, 180),
					makeTrackfile(4, 180),
				},
			},
			expected: brevity.UnknownDirection,
		},
		{
			name:     "empty group",
			group:    group{},
			expected: brevity.UnknownDirection,
		},
		{
			name: "furball",
			group: group{
				contacts:    []*trackfiles.Trackfile{makeTrackfile(1, 0)},
				declaration: brevity.Furball,
			},
			expected: brevity.UnknownDirection,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.group.Track())
		})
	}
}
