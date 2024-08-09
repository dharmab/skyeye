// package trackfiles records aircraft movement over time.
package trackfiles

import (
	"fmt"
	"math"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// Labels are identifying information attached to a trackfile.
type Labels struct {
	// UnitID is the in-game ID.
	UnitID uint32
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
	// Track is a collection of frames, ordered from most recent to least recent.
	Track deque.Deque[Frame]
	// MaxLength is the maximum number of frames to keep in the trackfile.
	// At least 3 are needed to distinguish linear tracks and curved tracks. More than 5 is probably excessive.
	MaxLength int
}

// Frame describes a contact's position and velocity at a point in time.
type Frame struct {
	// Time within the simulation when the event occurred.
	// This is not a wall-clock time. It may be decades in the past or years in the future, relative to the system clock.
	Time time.Time
	// Point is the contact's 2D position.
	Point orb.Point
	// Altitude above sea level.
	Altitude unit.Length
	// Heading is the direction the contact is moving. This is not necessarily the direction the nose is poining.
	Heading unit.Angle
}

func NewTrackfile(a Labels) *Trackfile {
	return &Trackfile{
		Contact:   a,
		Track:     *deque.New[Frame](),
		MaxLength: 4,
	}
}

func (t *Trackfile) String() string {
	point := t.LastKnown().Point
	return fmt.Sprintf(
		"%d %s (%f.3, %f.3) %f.0 ft %f.0 kts %q",
		t.Contact.UnitID,
		t.Contact.ACMIName,
		point.Lon(),
		point.Lat(),
		t.LastKnown().Altitude.Feet(),
		t.Speed().Knots(),
		t.Contact.Name,
	)
}

// Update the trackfile with a new frame. Frames older than the most recent one are discarded.
func (t *Trackfile) Update(f Frame) {
	if t.Track.Len() > 0 && f.Time.Before(t.Track.Front().Time) {
		return
	}
	t.Track.PushFront(f)
	for t.Track.Len() > t.MaxLength {
		t.Track.PopBack()
	}
}

// Bullseye returns the bearing and distance from the bullseye to the track's last known position.
func (t *Trackfile) Bullseye(bullseye orb.Point) brevity.Bullseye {
	latest := t.Track.Front()
	declination, _ := bearings.Declination(bullseye, latest.Time)
	bearing := bearings.NewTrueBearing(
		unit.Angle(
			geo.Bearing(bullseye, latest.Point),
		) * unit.Degree,
	).Magnetic(declination)
	log.Debug().Float64("bearing", bearing.Degrees()).Msg("calculated bullseye bearing for group")
	distance := unit.Length(geo.Distance(bullseye, latest.Point)) * unit.Meter
	return *brevity.NewBullseye(bearing, distance)
}

// LastKnown returns the most recent frame in the trackfile.
// If the trackfile is empty, a stub frame with a zero-value time is returned.
func (t *Trackfile) LastKnown() Frame {
	if t.Track.Len() == 0 {
		return Frame{}
	}
	return t.Track.Front()
}

func (t *Trackfile) bestAvailableDeclination() unit.Angle {
	declincation, err := bearings.Declination(t.LastKnown().Point, t.LastKnown().Time)
	if err != nil {
		return 0
	}
	return declincation
}

// Course returns the angle that the track is moving in.
// If the track has not moved very far, the course may be unreliable.
// You can check for this condition by checking if [Trackfile.Direction] returns [brevity.UnknownDirection].
func (t *Trackfile) Course() bearings.Bearing {
	if t.Track.Len() < 2 {
		return bearings.NewTrueBearing(
			unit.Angle(
				t.LastKnown().Heading,
			) * unit.Degree,
		).Magnetic(t.bestAvailableDeclination())
	}

	latest := t.Track.Front()
	previous := t.Track.At(1)

	course := bearings.NewTrueBearing(
		unit.Angle(
			geo.Bearing(previous.Point, latest.Point),
		) * unit.Degree,
	).Magnetic(t.bestAvailableDeclination())
	return course
}

// Direction returns the cardinal direction that the track is moving in, or [brevity.UnknownDirection] if the track is not moving faster than 1 m/s.
func (t *Trackfile) Direction() brevity.Track {
	if t.Track.Len() < 2 {
		return brevity.UnknownDirection
	}
	if t.groundSpeed() < 1*unit.MetersPerSecond {
		return brevity.UnknownDirection
	}

	course := t.Course()
	return brevity.TrackFromBearing(course)
}

// groundSpeed returns the approxmiate speed of the track along the ground (i.e. in two dimensions).
func (t *Trackfile) groundSpeed() unit.Speed {
	if t.Track.Len() < 2 {
		return 0
	}

	latest := t.Track.Front()
	previous := t.Track.At(1)

	timeDelta := latest.Time.Sub(previous.Time) + 1*time.Millisecond

	groundDistance := unit.Length(math.Abs(geo.Distance(latest.Point, previous.Point))) * unit.Meter
	groundSpeed := unit.Speed(
		groundDistance.Meters()/
			timeDelta.Seconds(),
	) * unit.MetersPerSecond

	return groundSpeed
}

// Speed returns either the ground speed or the true 3D speed of the track, whichever is greater.
func (t *Trackfile) Speed() unit.Speed {
	if t.Track.Len() < 2 {
		return 0
	}

	latest := t.Track.Front()
	previous := t.Track.At(1)
	timeDelta := latest.Time.Sub(previous.Time) + 1*time.Millisecond
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
