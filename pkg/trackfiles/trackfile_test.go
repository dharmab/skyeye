package trackfiles

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

func TestTracking(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name                 string
		heading              unit.Angle
		ΔX                   unit.Length
		ΔY                   unit.Length
		ΔZ                   unit.Length
		ΔT                   time.Duration
		expectedApproxCourse unit.Angle
		expectedDirection    brevity.Track
		expectedApproxSpeed  unit.Speed
	}{
		{
			name:                 "North",
			heading:              0 * unit.Degree,
			ΔX:                   0 * unit.Meter,
			ΔY:                   200 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 0 * unit.Degree,
			expectedDirection:    brevity.North,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "Northeast",
			heading:              45 * unit.Degree,
			ΔX:                   100 * unit.Meter,
			ΔY:                   100 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 45 * unit.Degree,
			expectedDirection:    brevity.Northeast,
			expectedApproxSpeed:  70.7 * unit.MetersPerSecond,
		},
		{
			name:                 "East",
			heading:              90 * unit.Degree,
			ΔX:                   200 * unit.Meter,
			ΔY:                   0 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 90 * unit.Degree,
			expectedDirection:    brevity.East,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "Southeast",
			heading:              135 * unit.Degree,
			ΔX:                   100 * unit.Meter,
			ΔY:                   -100 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 135 * unit.Degree,
			expectedDirection:    brevity.Southeast,
			expectedApproxSpeed:  70.7 * unit.MetersPerSecond,
		},
		{
			name:                 "South",
			heading:              180 * unit.Degree,
			ΔX:                   0 * unit.Meter,
			ΔY:                   -200 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 180 * unit.Degree,
			expectedDirection:    brevity.South,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "Southwest",
			heading:              225 * unit.Degree,
			ΔX:                   -100 * unit.Meter,
			ΔY:                   -100 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 225 * unit.Degree,
			expectedDirection:    brevity.Southwest,
			expectedApproxSpeed:  70.7 * unit.MetersPerSecond,
		},
		{
			name:                 "West",
			heading:              270 * unit.Degree,
			ΔX:                   -200 * unit.Meter,
			ΔY:                   0 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 270 * unit.Degree,
			expectedDirection:    brevity.West,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "Northwest",
			heading:              315 * unit.Degree,
			ΔX:                   -100 * unit.Meter,
			ΔY:                   100 * unit.Meter,
			ΔZ:                   0 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 315 * unit.Degree,
			expectedDirection:    brevity.Northwest,
			expectedApproxSpeed:  70.7 * unit.MetersPerSecond,
		},
		{
			name:                 "Vertical climb",
			heading:              0 * unit.Degree,
			ΔX:                   0 * unit.Meter,
			ΔY:                   0 * unit.Meter,
			ΔZ:                   200 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 0 * unit.Degree,
			expectedDirection:    brevity.UnknownDirection,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "Vertical dive",
			heading:              0 * unit.Degree,
			ΔX:                   0 * unit.Meter,
			ΔY:                   0 * unit.Meter,
			ΔZ:                   -200 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 0 * unit.Degree,
			expectedDirection:    brevity.UnknownDirection,
			expectedApproxSpeed:  100 * unit.MetersPerSecond,
		},
		{
			name:                 "3D motion",
			heading:              45 * unit.Degree,
			ΔX:                   100 * unit.Meter,
			ΔY:                   100 * unit.Meter,
			ΔZ:                   100 * unit.Meter,
			ΔT:                   2 * time.Second,
			expectedApproxCourse: 45 * unit.Degree,
			expectedDirection:    brevity.Northeast,
			expectedApproxSpeed:  86.6 * unit.MetersPerSecond,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			trackfile := NewTrackfile(Labels{
				ID:        1,
				ACMIName:  "F-15C",
				Name:      "Eagle 1",
				Coalition: coalitions.Blue,
			})
			now := time.Now()
			alt := 20000 * unit.Foot

			trackfile.Update(Frame{
				Time:     now.Add(-1 * test.ΔT),
				Point:    orb.Point{-115.0338, 36.2350},
				Altitude: alt,
				Heading:  test.heading,
			})
			dest := spatial.PointAtBearingAndDistance(trackfile.LastKnown().Point, bearings.NewTrueBearing(0), test.ΔY)
			dest = spatial.PointAtBearingAndDistance(dest, bearings.NewTrueBearing(90*unit.Degree), test.ΔX)
			trackfile.Update(Frame{
				Time:     now,
				Point:    dest,
				Altitude: alt + test.ΔZ,
				Heading:  test.heading,
			})

			require.InDelta(t, test.expectedApproxSpeed.MetersPerSecond(), trackfile.Speed().MetersPerSecond(), 0.5)
			require.Equal(t, test.expectedDirection, trackfile.Direction())
			if test.expectedDirection != brevity.UnknownDirection {
				declination, err := bearings.Declination(dest, now)
				require.NoError(t, err)
				require.InDelta(t, bearings.NewTrueBearing(test.expectedApproxCourse).Magnetic(declination).Degrees(), trackfile.Course().Degrees(), 0.5)
			}
		})
	}
}
