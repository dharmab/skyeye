package radar

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestNearest(t *testing.T) {
	testCases := []struct {
		name       string
		origin     orb.Point
		pointA     orb.Point
		pointB     orb.Point
		expectedID uint64
	}{
		{
			name:       "finds nearest Red aircraft to origin",
			origin:     orb.Point{33.405794, 69.047461},
			pointA:     orb.Point{33.405794, 69.047461},
			pointB:     orb.Point{24.973478, 70.068836},
			expectedID: 2,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			starts := make(chan sim.Started, 10)
			updates := make(chan sim.Updated, 10)
			fades := make(chan sim.Faded, 10)
			rdr := New(coalitions.Blue, starts, updates, fades, 20*unit.NauticalMile)
			rdr.SetMissionTime(time.Date(1999, 06, 11, 12, 0, 0, 0, time.UTC))
			rdr.SetBullseye(orb.Point{22.867128, 69.047461}, coalitions.Blue)

			// Start the radar to process updates
			ctx, cancel := context.WithCancel(context.Background())
			wg := &sync.WaitGroup{}
			go rdr.Run(ctx, wg)

			// Add trackfiles to radar via updates channel
			updates <- sim.Updated{
				Labels: trackfiles.Labels{
					ID:        1,
					ACMIName:  "F-15C",
					Name:      "Eagle 1",
					Coalition: coalitions.Blue,
				},
				Frame: trackfiles.Frame{
					Time:     time.Date(1999, 06, 11, 12, 0, 0, 0, time.UTC),
					Point:    test.pointA,
					Altitude: 30000 * unit.Foot,
					AGL: func() *unit.Length {
						agl := 30000 * unit.Foot
						return &agl
					}(),
					Heading: 90 * unit.Degree,
				},
			}
			updates <- sim.Updated{
				Labels: trackfiles.Labels{
					ID:        2,
					ACMIName:  "F-15C",
					Name:      "Eagle 2",
					Coalition: coalitions.Red,
				},
				Frame: trackfiles.Frame{
					Time:     time.Date(1999, 06, 11, 12, 0, 0, 0, time.UTC),
					Point:    test.pointB,
					Altitude: 30000 * unit.Foot,
					AGL: func() *unit.Length {
						agl := 30000 * unit.Foot
						return &agl
					}(),
					Heading: 90 * unit.Degree,
				},
			}

			// Wait for updates to be processed
			time.Sleep(100 * time.Millisecond)

			group := rdr.FindNearestGroupWithBRAA(
				test.origin,
				0*unit.NauticalMile,
				100000*unit.Foot,
				300*unit.NauticalMile,
				coalitions.Red,
				0,
			)

			assert.NotNil(t, group)
			if group != nil {
				ids := group.ObjectIDs()
				assert.Contains(t, ids, test.expectedID)
			}

			// Clean up
			cancel()
		})
	}
}
