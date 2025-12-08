package trackfiles

import (
	"fmt"
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	spatial.ForceTerrain("Kola", spatial.KolaProjection())
}

func TestTracking(t *testing.T) {
	t.Parallel()
	spatial.ResetTerrainToDefault()
	spatial.ForceTerrain("Kola", spatial.KolaProjection())
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
			trackfile := New(Labels{
				ID:        1,
				ACMIName:  "F-15C",
				Name:      "Eagle 1",
				Coalition: coalitions.Blue,
			})
			//now := time.Now()
			now := time.Date(1999, 06, 11, 12, 0, 0, 0, time.UTC)

			alt := 20000 * unit.Foot

			trackfile.Update(Frame{
				Time:     now.Add(-1 * test.ΔT),
				Point:    orb.Point{33.405794, 69.047461},
				Altitude: alt,
				Heading:  test.heading,
			})
			dest := spatial.PointAtBearingAndDistance(trackfile.LastKnown().Point, bearings.NewTrueBearing(0), test.ΔY) // translate point in Y axis
			dest = spatial.PointAtBearingAndDistance(dest, bearings.NewTrueBearing(90*unit.Degree), test.ΔX)            // translate point in X axis
			trackfile.Update(Frame{
				Time:     now,
				Point:    dest,
				Altitude: alt + test.ΔZ,
				Heading:  test.heading,
			})

			assert.InDelta(t, test.expectedApproxSpeed.MetersPerSecond(), trackfile.Speed().MetersPerSecond(), 1)
			assert.Equal(t, test.expectedDirection, trackfile.Direction())
			if test.expectedDirection != brevity.UnknownDirection {
				declination, err := bearings.Declination(dest, now)
				//fmt.Printf("declination at %f,%f is %f\n", dest.Lat(), dest.Lon(), declination.Degrees())
				require.NoError(t, err)
				//fmt.Printf("NewTrueBearing(test.expectedApproxCourse) %f\n", bearings.NewTrueBearing(test.expectedApproxCourse).Degrees())
				//fmt.Printf("NewTrueBearing(test.expectedApproxCourse).Magnetic(declination) %f\n", bearings.NewTrueBearing(test.expectedApproxCourse).Magnetic(declination).Degrees())
				//fmt.Printf("trackfile.Course() %f\n", trackfile.Course().Degrees())

				//fmt.Printf("trackfile.Speed() %f\n", trackfile.Speed().MetersPerSecond())

				assert.InDelta(t, bearings.NewTrueBearing(test.expectedApproxCourse).Magnetic(declination).Degrees(), trackfile.Course().Degrees(), 0.5)
			}
		})
	}
}

func TestBullseye(t *testing.T) { // tests bullseye calculations - bearing and distance to trackfile point given bullseye point
	t.Parallel()
	spatial.ResetTerrainToDefault()
	spatial.ForceTerrain("Kola", spatial.KolaProjection())
	trackfile := New(Labels{
		ID:        1,
		ACMIName:  "F-15C",
		Name:      "Eagle 1",
		Coalition: coalitions.Blue,
	}) //		target:           orb.Point{33.405794, 69.047461},
	now := time.Date(1999, 06, 11, 12, 0, 0, 0, time.UTC)
	alt := 20000 * unit.Foot
	heading := 0 * unit.Degree
	testCases := []struct {
		bullseye         orb.Point
		expectedBearing  unit.Angle
		expectedDistance unit.Length
		tf_point         orb.Point
	}{
		{
			bullseye:         orb.Point{22.867128, 68.474419},
			tf_point:         orb.Point{33.405794, 69.047461}, // kola Sveromorsk-1
			expectedBearing:  62 * unit.Degree,                // magnetic
			expectedDistance: 232 * unit.NauticalMile,
		},
		{
			bullseye:         orb.Point{22.867128, 68.474419},
			tf_point:         orb.Point{24.973478, 70.068836}, // kola Banak
			expectedBearing:  14 * unit.Degree,                // magnetic
			expectedDistance: 106 * unit.NauticalMile,
		},
		{
			bullseye:         orb.Point{22.867128, 68.474419},
			tf_point:         orb.Point{34.262989, 64.91865}, // kola Poduzhemye
			expectedBearing:  110 * unit.Degree,              // magnetic
			expectedDistance: 345 * unit.NauticalMile,
		},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v -> %v", test.bullseye, test.tf_point), func(t *testing.T) {
			t.Parallel()
			trackfile.Update(Frame{
				Time:     now,
				Point:    test.tf_point,
				Altitude: alt,
				Heading:  heading,
			})
			actual := trackfile.Bullseye(test.bullseye)
			assert.InDelta(t, test.expectedDistance.NauticalMiles(), actual.Distance().NauticalMiles(), 5)
			assert.InDelta(t, test.expectedBearing.Degrees(), actual.Bearing().Degrees(), 5)
		})
	}
}
