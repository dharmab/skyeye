// Package trackfiles records aircraft movement over time.
package trackfiles

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
)

// Labels are identifying information attached to a trackfile.
type Labels struct {
	// ID is the object ID from TacView.
	ID uint64
	// Name is a unique string for each aircraft.
	// If this is a player's aircraft, it will be the player's in-game name.
	Name string
	// Coalition the aircraft belongs to.
	Coalition coalitions.Coalition
	// The name of the aircraft type in the ACMI file.
	// See https://www.tacview.net/documentation/database/en/
	ACMIName string
}

// Trackfile tracks a contact's movement over time.
type Trackfile struct {
	// Contact contains identifying information.
	Contact Labels
	// track is a collection of frames, ordered from most recent to least recent.
	track deque.Deque[Frame]
	lock  sync.RWMutex
}

const maxLength = 4

// Frame describes a contact's position and velocity at a point in time.
type Frame struct {
	// Time within the simulation when the event occurred.
	// This is not a wall-clock time. It may be decades in the past or years in the future, relative to the system clock.
	Time time.Time
	// Point is the contact's 2D position.
	Point orb.Point
	// Altitude above sea level.
	Altitude unit.Length
	// Altitude above ground level, if available.
	AGL *unit.Length
	// Heading is the direction the contact is moving. This is not necessarily the direction the nose is pointing.
	Heading unit.Angle
}

// New creates a new trackfile with the given labels.
func New(labels Labels) *Trackfile {
	return &Trackfile{
		Contact: labels,
		track:   *deque.New[Frame](),
	}
}

// String returns a string representation of the trackfile suitable for logging.
func (t *Trackfile) String() string {
	frame := t.LastKnown()
	return fmt.Sprintf(
		"%d %s (%f.3, %f.3) %f.0 ft %f.0 kts %q",
		t.Contact.ID,
		t.Contact.ACMIName,
		frame.Point.Lon(),
		frame.Point.Lat(),
		frame.Altitude.Feet(),
		t.Speed().Knots(),
		t.Contact.Name,
	)
}

// Update the trackfile with a new frame. Frames older than the most recent one are discarded.
func (t *Trackfile) Update(f Frame) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.track.Len() > 0 && f.Time.Before(t.track.Front().Time) {
		return
	}
	t.track.PushFront(f)
	for t.track.Len() > maxLength {
		t.track.PopBack()
	}
}

// Bullseye returns the bearing and distance from the bullseye to the track's last known position.
func (t *Trackfile) Bullseye(bullseye orb.Point) brevity.Bullseye {
	latest := t.LastKnown()
	declination, _ := bearings.Declination(latest.Point, latest.Time)
	//log.Debug().Any("declination", declination.Degrees()).Msgf("computed magnetic trackfilebullseye declination at point")

	bearing := spatial.TrueBearing(bullseye, latest.Point).Magnetic(declination)
	//log.Debug().Float64("bearing", bearing.Degrees()).Msg("calculated bullseye bearing for group")
	distance := spatial.Distance(bullseye, latest.Point)
	return *brevity.NewBullseye(bearing, distance)
}

// LastKnown returns the most recent frame in the trackfile.
// If the trackfile is empty, a stub frame with a zero-value time is returned.
func (t *Trackfile) LastKnown() Frame {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.unsafeLastKnown()
}

// unsafeLastKnown is like LastKnown, but it does not acquire a lock. The calling
// function must acquire t.lock before calling this function.
func (t *Trackfile) unsafeLastKnown() Frame {
	if t.track.Len() == 0 {
		return Frame{}
	}
	return t.track.Front()
}

// IsLastKnownPointZero returns true if the last known point is at (0, 0).
// This means the trackfile has recorded no data.
func (t *Trackfile) IsLastKnownPointZero() bool {
	return spatial.IsZero(t.LastKnown().Point)
}

func (t *Trackfile) bestAvailableDeclination() unit.Angle {
	latest := t.unsafeLastKnown()
	declination, err := bearings.Declination(latest.Point, latest.Time)
	//log.Debug().Any("declination", declination).Msgf("computed bestAvailableDeclination magnetic declination at point lat %f lon %f", latest.Point.Lat(), latest.Point.Lon())

	if err != nil {
		return 0
	}
	return declination
}

// Course returns the angle that the track is moving in, relative to magnetic north.
// If the track has not moved very far, the course may be unreliable.
// You can check for this condition by checking if [Trackfile.Direction] returns [brevity.UnknownDirection].
func (t *Trackfile) Course() bearings.Bearing {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.track.Len() == 1 {
		return bearings.NewTrueBearing(
			unit.Angle(
				t.track.Front().Heading,
			) * unit.Degree,
		).Magnetic(t.bestAvailableDeclination())
	}

	latest := t.track.Front()
	previous := t.track.At(1)

	declination := t.bestAvailableDeclination()
	course := spatial.TrueBearing(previous.Point, latest.Point).Magnetic(declination)
	return course
}

// Direction returns the cardinal direction that the track is moving in, or [brevity.UnknownDirection] if the track is not moving faster than 1 m/s.
func (t *Trackfile) Direction() brevity.Track {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.track.Len() < 2 {
		return brevity.UnknownDirection
	}
	if t.groundSpeed() < 1*unit.MetersPerSecond {
		return brevity.UnknownDirection
	}

	course := t.Course()
	return brevity.TrackFromBearing(course)
}

// groundSpeed returns the approximate speed of the track along the ground (i.e. in two dimensions).
func (t *Trackfile) groundSpeed() unit.Speed {
	if t.track.Len() < 2 {
		return 0
	}

	latest := t.track.Front()
	previous := t.track.At(1)

	timeDelta := latest.Time.Sub(previous.Time)
	if timeDelta == 0 {
		return 0
	}

	groundDistance := spatial.Distance(latest.Point, previous.Point)
	groundSpeed := unit.Speed(
		groundDistance.Meters()/
			timeDelta.Seconds(),
	) * unit.MetersPerSecond

	return groundSpeed
}

// Speed returns either the ground speed or the true 3D speed of the track, whichever is greater.
func (t *Trackfile) Speed() unit.Speed {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.track.Len() < 2 {
		return 0
	}

	latest := t.track.Front()
	previous := t.track.At(1)

	timeDelta := latest.Time.Sub(previous.Time)
	if timeDelta == 0 {
		return 0
	}

	var verticalDistance unit.Length
	if latest.Altitude > previous.Altitude {
		verticalDistance = latest.Altitude - previous.Altitude
	} else {
		verticalDistance = previous.Altitude - latest.Altitude
	}
	verticalSpeed := unit.Speed(
		verticalDistance.Meters()/
			timeDelta.Seconds(),
	) * unit.MetersPerSecond

	groundSpeed := t.groundSpeed()

	trueSpeed := unit.Speed(
		math.Sqrt(
			math.Pow(groundSpeed.MetersPerSecond(), 2)+math.Pow(verticalSpeed.MetersPerSecond(), 2),
		),
	) * unit.MetersPerSecond

	if groundSpeed > trueSpeed {
		return groundSpeed
	}
	return trueSpeed
}
