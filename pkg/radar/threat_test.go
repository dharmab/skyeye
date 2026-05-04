package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

// makeReceiverAtOffset builds a trackfile positioned at the given true bearing and distance from
// an origin. The bearing is in degrees and the distance in nautical miles.
func makeReceiverAtOffset(id uint64, origin orb.Point, trueBearingDegrees float64, rangeNM float64) *trackfiles.Trackfile {
	bearing := bearings.NewTrueBearing(unit.Angle(trueBearingDegrees) * unit.Degree)
	point := spatial.PointAtBearingAndDistance(origin, bearing, unit.Length(rangeNM)*unit.NauticalMile)
	return makeTrackfileAt(id, point)
}

func makeTrackfileAt(id uint64, point orb.Point) *trackfiles.Trackfile {
	tf := trackfiles.New(trackfiles.Labels{
		ID:        id,
		ACMIName:  "F-15C",
		Name:      "Eagle",
		Coalition: coalitions.Blue,
	})
	tf.Update(trackfiles.Frame{
		Time:     time.Now(),
		Point:    point,
		Altitude: 20000 * unit.Foot,
	})
	return tf
}

func TestSharedBRAAOrigin(t *testing.T) {
	t.Parallel()
	// Place the hostile 40nm due north of a common anchor. Receivers are placed around the
	// anchor so we can reason about their BRAAs to the hostile easily.
	anchor := orb.Point{-115.0, 36.0}
	hostilePoint := spatial.PointAtBearingAndDistance(
		anchor,
		bearings.NewTrueBearing(0),
		40*unit.NauticalMile,
	)
	hostile := &group{contacts: []*trackfiles.Trackfile{makeTrackfileAt(100, hostilePoint)}}

	testCases := []struct {
		name      string
		receivers []*trackfiles.Trackfile
		want      bool
	}{
		{
			name:      "no receivers",
			receivers: nil,
			want:      false,
		},
		{
			name: "single receiver",
			receivers: []*trackfiles.Trackfile{
				makeTrackfileAt(1, anchor),
			},
			want: false,
		},
		{
			name: "two receivers colocated",
			receivers: []*trackfiles.Trackfile{
				makeTrackfileAt(1, anchor),
				makeReceiverAtOffset(2, anchor, 90, 0.1),
			},
			want: true,
		},
		{
			name: "bearing spread too wide",
			receivers: []*trackfiles.Trackfile{
				makeTrackfileAt(1, anchor),
				// Offset east by 10nm — bearing from receiver to hostile shifts by > 5 degrees.
				makeReceiverAtOffset(2, anchor, 90, 10),
			},
			want: false,
		},
		{
			name: "range spread too wide",
			receivers: []*trackfiles.Trackfile{
				makeTrackfileAt(1, anchor),
				// Offset south by 5nm — range to hostile grows by ~5nm, which exceeds 1nm.
				makeReceiverAtOffset(2, anchor, 180, 5),
			},
			want: false,
		},
	}

	r := &Radar{
		maxSharedBRAABearingSpread: 5 * unit.Degree,
		maxSharedBRAARangeSpread:   1 * unit.NauticalMile,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, ok := r.getGroupBRAAOrigin(hostile, tc.receivers)
			assert.Equal(t, tc.want, ok)
		})
	}
}
