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
	"github.com/stretchr/testify/require"
)

func TestFindNearbyGroupsStableDistanceOrdering(t *testing.T) {
	t.Parallel()

	radar := New(coalitions.Blue, nil, nil, nil, 0, false)
	pointOfInterest := orb.Point{-115, 36}
	agl := 100 * unit.Meter
	now := time.Now()

	makeTrackfile := func(id uint64, distance unit.Length, bearing bearings.Bearing) *trackfiles.Trackfile {
		tf := trackfiles.New(trackfiles.Labels{
			ID:        id,
			Name:      "Test",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		})
		tf.Update(trackfiles.Frame{
			Time:     now,
			Point:    spatial.PointAtBearingAndDistance(pointOfInterest, bearing, distance),
			Altitude: 20000 * unit.Foot,
			AGL:      &agl,
		})
		return tf
	}

	radar.contacts.set(makeTrackfile(1, 50*unit.Kilometer, bearings.NewTrueBearing(90*unit.Degree)))
	radar.contacts.set(makeTrackfile(2, 50*unit.Kilometer+0.5*unit.Meter, bearings.NewTrueBearing(180*unit.Degree)))

	for range 20 {
		groups := radar.FindNearbyGroupsWithBullseye(
			pointOfInterest,
			0,
			50000*unit.Foot,
			100*unit.Kilometer,
			coalitions.Blue,
			brevity.Aircraft,
			nil,
		)
		require.Len(t, groups, 2)
		assert.Equal(t, uint64(1), groups[0].ObjectIDs()[0])
		assert.Equal(t, uint64(2), groups[1].ObjectIDs()[0])
	}
}
