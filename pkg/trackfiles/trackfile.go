// Package trackfiles records aircraft movement over time.
package trackfiles

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dharmab/collections/deques"
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
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
	track *deques.Counting[Frame]
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
	// Heading is the direction the contact is moving. This is not necessarily the direction the nose is poining.
	Heading unit.Angle
}

// New creates a new trackfile with the given labels.
func New(labels Labels) *Trackfile {
	return &Trackfile{
		Contact: labels,
		track:   deques.NewCounting[Frame](maxLength),
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
	if newest, ok := t.track.Newest(); ok && f.Time.Before(newest.Time) {
		return
	}
	t.track.Push(f)
}

// Bullseye returns the bearing and distance from the bullseye to the track's last known position.
func (t *Trackfile) Bullseye(bullseye orb.Point, opts ...spatial.Option) *brevity.Bullseye {
	latest := t.LastKnown()
	declination, _ := bearings.Declination(bullseye, latest.Time)
	bearing := spatial.TrueBearing(bullseye, latest.Point, opts...).Magnetic(declination)
	log.Debug().Float64("bearing", bearing.Degrees()).Msg("calculated bullseye bearing for group")
	distance := spatial.Distance(bullseye, latest.Point, opts...)
	return brevity.NewBullseye(bearing, distance)
}

// LastKnown returns the most recent frame in the trackfile.
// If the trackfile is empty, a stub frame with a zero-value time is returned.
func (t *Trackfile) LastKnown() Frame {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if f, ok := t.track.Newest(); ok {
		return f
	}
	return Frame{}
}

// IsLastKnownPointZero returns true if the last known point is at (0, 0).
// This means the trackfile has recorded no data.
func (t *Trackfile) IsLastKnownPointZero() bool {
	return spatial.IsZero(t.LastKnown().Point)
}

// bestAvailableDeclination returns the magnetic declination at the track's most recent position, or 0 if unavailable.
func (t *Trackfile) bestAvailableDeclination() unit.Angle {
	latest, ok := t.track.Newest()
	if !ok {
		return 0
	}
	declincation, err := bearings.Declination(latest.Point, latest.Time)
	if err != nil {
		return 0
	}
	return declincation
}

// Course returns the angle that the track is moving in.
// If the track has not moved very far, the course may be unreliable.
// You can check for this condition by checking if [Trackfile.Direction] returns [brevity.UnknownDirection].
func (t *Trackfile) Course(opts ...spatial.Option) bearings.Bearing {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.computeCourse(opts...)
}

// computeCourse returns the magnetic bearing between the two most recent frames, or the heading if only one frame exists. Caller must hold t.lock.
func (t *Trackfile) computeCourse(opts ...spatial.Option) bearings.Bearing {
	latest, ok := t.track.Newest()
	if !ok {
		return bearings.NewTrueBearing(0)
	}
	if t.track.Len() == 1 {
		return bearings.NewTrueBearing(
			latest.Heading,
		).Magnetic(t.bestAvailableDeclination())
	}

	previous, _ := t.track.At(1)

	declination := t.bestAvailableDeclination()
	course := spatial.TrueBearing(previous.Point, latest.Point, opts...).Magnetic(declination)
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

	return brevity.TrackFromBearing(t.computeCourse())
}

// groundSpeed returns the approximate ground speed of the track in two dimensions. Caller must hold t.lock.
func (t *Trackfile) groundSpeed(opts ...spatial.Option) unit.Speed {
	latest, ok := t.track.Newest()
	if !ok {
		return 0
	}
	previous, ok := t.track.At(1)
	if !ok {
		return 0
	}

	timeDelta := latest.Time.Sub(previous.Time)
	if timeDelta == 0 {
		return 0
	}

	groundDistance := spatial.Distance(latest.Point, previous.Point, opts...)
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
	latest, ok := t.track.Newest()
	if !ok {
		return 0
	}
	previous, ok := t.track.At(1)
	if !ok {
		return 0
	}

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
