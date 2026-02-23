package trackfiles

import (
	"sync"
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

var testLabels = Labels{
	ID:        1,
	ACMIName:  "F-15C",
	Name:      "Eagle 1",
	Coalition: coalitions.Blue,
}

var testOrigin = orb.Point{-115.0338, 36.2350}

func TestNew(t *testing.T) {
	t.Parallel()
	tf := New(testLabels)
	require.NotNil(t, tf)
	assert.Equal(t, testLabels.ID, tf.Contact.ID)
	assert.Equal(t, testLabels.ACMIName, tf.Contact.ACMIName)
	assert.Equal(t, testLabels.Name, tf.Contact.Name)
	assert.Equal(t, testLabels.Coalition, tf.Contact.Coalition)
	assert.Equal(t, Frame{}, tf.LastKnown())
}

func TestLastKnown(t *testing.T) {
	t.Parallel()
	t.Run("empty trackfile returns zero frame", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		frame := tf.LastKnown()
		assert.True(t, frame.Time.IsZero())
		assert.True(t, spatial.IsZero(frame.Point))
	})
	t.Run("single update returns that frame", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{
			Time:     now,
			Point:    testOrigin,
			Altitude: 20000 * unit.Foot,
			Heading:  90 * unit.Degree,
		})
		frame := tf.LastKnown()
		assert.Equal(t, now, frame.Time)
		assert.Equal(t, testOrigin, frame.Point)
	})
	t.Run("multiple updates returns most recent", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now.Add(-2 * time.Second), Point: testOrigin})
		tf.Update(Frame{Time: now.Add(-1 * time.Second), Point: testOrigin})
		latest := Frame{Time: now, Point: orb.Point{-115.0, 36.0}, Altitude: 30000 * unit.Foot}
		tf.Update(latest)
		frame := tf.LastKnown()
		assert.Equal(t, latest.Time, frame.Time)
		assert.Equal(t, latest.Point, frame.Point)
		assert.InDelta(t, latest.Altitude.Feet(), frame.Altitude.Feet(), 1)
	})
}

func TestIsLastKnownPointZero(t *testing.T) {
	t.Parallel()
	t.Run("new trackfile is zero", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		assert.True(t, tf.IsLastKnownPointZero())
	})
	t.Run("origin (0,0) is zero", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		tf.Update(Frame{Time: time.Now(), Point: orb.Point{0, 0}})
		assert.True(t, tf.IsLastKnownPointZero())
	})
	t.Run("non-zero point is not zero", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		tf.Update(Frame{Time: time.Now(), Point: testOrigin})
		assert.False(t, tf.IsLastKnownPointZero())
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	t.Run("out-of-order frames are discarded", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now, Point: testOrigin, Altitude: 20000 * unit.Foot})
		// Older frame should be discarded
		tf.Update(Frame{Time: now.Add(-5 * time.Second), Point: orb.Point{-116.0, 37.0}, Altitude: 10000 * unit.Foot})
		frame := tf.LastKnown()
		assert.Equal(t, now, frame.Time)
		assert.Equal(t, testOrigin, frame.Point)
	})
	t.Run("frame limit enforced", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		// Add 6 frames; only the 4 most recent should be retained
		for i := range 6 {
			tf.Update(Frame{
				Time:     now.Add(time.Duration(i) * time.Second),
				Point:    testOrigin,
				Altitude: unit.Length(i*1000) * unit.Foot,
			})
		}
		// LastKnown should be the most recent (i=5)
		frame := tf.LastKnown()
		assert.Equal(t, now.Add(5*time.Second), frame.Time)
		// Speed should still work (uses two most recent frames)
		assert.NotZero(t, tf.Speed())
	})
	t.Run("sequential updates all accepted", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		for i := range 4 {
			tf.Update(Frame{
				Time:     now.Add(time.Duration(i) * time.Second),
				Point:    testOrigin,
				Altitude: unit.Length(20000+i*100) * unit.Foot,
			})
		}
		frame := tf.LastKnown()
		assert.Equal(t, now.Add(3*time.Second), frame.Time)
	})
}

func TestCourse(t *testing.T) {
	t.Parallel()
	t.Run("single frame uses heading field", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		heading := 135 * unit.Degree
		tf.Update(Frame{
			Time:    now,
			Point:   testOrigin,
			Heading: heading,
		})
		declination, err := bearings.Declination(testOrigin, now)
		require.NoError(t, err)
		expected := bearings.NewTrueBearing(heading).Magnetic(declination)
		assert.InDelta(t, expected.Degrees(), tf.Course().Degrees(), 0.5)
	})
	t.Run("two frames calculates bearing between points", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now.Add(-2 * time.Second), Point: testOrigin})
		dest := spatial.PointAtBearingAndDistance(testOrigin, bearings.NewTrueBearing(90*unit.Degree), 200*unit.Meter)
		tf.Update(Frame{Time: now, Point: dest})

		declination, err := bearings.Declination(dest, now)
		require.NoError(t, err)
		expected := spatial.TrueBearing(testOrigin, dest).Magnetic(declination)
		assert.InDelta(t, expected.Degrees(), tf.Course().Degrees(), 0.5)
	})
}

func TestDirection(t *testing.T) {
	t.Parallel()
	t.Run("empty trackfile", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		assert.Equal(t, brevity.UnknownDirection, tf.Direction())
	})
	t.Run("single frame", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		tf.Update(Frame{Time: time.Now(), Point: testOrigin, Heading: 90 * unit.Degree})
		assert.Equal(t, brevity.UnknownDirection, tf.Direction())
	})
	t.Run("stationary", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now.Add(-2 * time.Second), Point: testOrigin, Altitude: 20000 * unit.Foot})
		tf.Update(Frame{Time: now, Point: testOrigin, Altitude: 20000 * unit.Foot})
		assert.Equal(t, brevity.UnknownDirection, tf.Direction())
	})
	t.Run("very slow movement", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now.Add(-2 * time.Second), Point: testOrigin})
		// Move less than 1 m/s (< 2 meters in 2 seconds)
		dest := spatial.PointAtBearingAndDistance(testOrigin, bearings.NewTrueBearing(0), 1*unit.Meter)
		tf.Update(Frame{Time: now, Point: dest})
		assert.Equal(t, brevity.UnknownDirection, tf.Direction())
	})
}

func TestSpeed(t *testing.T) {
	t.Parallel()
	t.Run("empty trackfile", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		assert.Zero(t, tf.Speed())
	})
	t.Run("single frame", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		tf.Update(Frame{Time: time.Now(), Point: testOrigin, Altitude: 20000 * unit.Foot})
		assert.Zero(t, tf.Speed())
	})
	t.Run("zero time delta", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		tf.Update(Frame{Time: now, Point: testOrigin})
		dest := spatial.PointAtBearingAndDistance(testOrigin, bearings.NewTrueBearing(90*unit.Degree), 200*unit.Meter)
		tf.Update(Frame{Time: now, Point: dest})
		assert.Zero(t, tf.Speed())
	})
	t.Run("horizontal movement", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		now := time.Now()
		alt := 20000 * unit.Foot
		tf.Update(Frame{Time: now.Add(-2 * time.Second), Point: testOrigin, Altitude: alt})
		dest := spatial.PointAtBearingAndDistance(testOrigin, bearings.NewTrueBearing(0), 200*unit.Meter)
		tf.Update(Frame{Time: now, Point: dest, Altitude: alt})
		assert.InDelta(t, 100.0, tf.Speed().MetersPerSecond(), 0.5)
	})
}

func TestBullseye(t *testing.T) {
	t.Parallel()
	tf := New(testLabels)
	now := time.Now()
	trackPoint := spatial.PointAtBearingAndDistance(testOrigin, bearings.NewTrueBearing(90*unit.Degree), 50*unit.NauticalMile)
	tf.Update(Frame{Time: now, Point: trackPoint, Altitude: 25000 * unit.Foot})

	bullseye := tf.Bullseye(testOrigin)
	// Distance should be approximately 50 NM
	assert.InDelta(t, 50.0, bullseye.Distance().NauticalMiles(), 1.0)

	// Bearing should be magnetic (approximately 90 true + declination)
	declination, err := bearings.Declination(testOrigin, now)
	require.NoError(t, err)
	expectedBearing := bearings.NewTrueBearing(90 * unit.Degree).Magnetic(declination)
	assert.InDelta(t, expectedBearing.Degrees(), bullseye.Bearing().Degrees(), 1.0)
}

func TestString(t *testing.T) {
	t.Parallel()
	t.Run("non-empty trackfile", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		tf.Update(Frame{Time: time.Now(), Point: testOrigin, Altitude: 20000 * unit.Foot})
		s := tf.String()
		assert.NotEmpty(t, s)
		assert.Contains(t, s, "F-15C")
		assert.Contains(t, s, "Eagle 1")
	})
	t.Run("empty trackfile does not panic", func(t *testing.T) {
		t.Parallel()
		tf := New(testLabels)
		assert.NotPanics(t, func() {
			s := tf.String()
			assert.NotEmpty(t, s)
		})
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Parallel()
	tf := New(testLabels)
	now := time.Now()
	const goroutines = 10
	const iterations = 100

	var wg sync.WaitGroup

	// Writers
	for i := range goroutines {
		wg.Go(func() {
			for j := range iterations {
				tf.Update(Frame{
					Time:     now.Add(time.Duration(i*iterations+j) * time.Millisecond),
					Point:    testOrigin,
					Altitude: unit.Length(20000+j) * unit.Foot,
					Heading:  unit.Angle(j%360) * unit.Degree,
				})
			}
		})
	}

	// Readers
	for range goroutines {
		wg.Go(func() {
			for range iterations {
				_ = tf.LastKnown()
				_ = tf.Speed()
				_ = tf.Direction()
				_ = tf.IsLastKnownPointZero()
			}
		})
	}

	wg.Wait()
}

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
			trackfile := New(Labels{
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

			assert.InDelta(t, test.expectedApproxSpeed.MetersPerSecond(), trackfile.Speed().MetersPerSecond(), 0.5)
			assert.Equal(t, test.expectedDirection, trackfile.Direction())
			if test.expectedDirection != brevity.UnknownDirection {
				declination, err := bearings.Declination(dest, now)
				require.NoError(t, err)
				assert.InDelta(t, bearings.NewTrueBearing(test.expectedApproxCourse).Magnetic(declination).Degrees(), trackfile.Course().Degrees(), 0.5)
			}
		})
	}
}
