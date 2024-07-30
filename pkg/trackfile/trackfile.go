package trackfile

import (
	"fmt"
	"math"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

type Aircraft struct {
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
	Contact Aircraft
	// Track is a collection of frames, ordered from most recent to least recent.
	Track deque.Deque[Frame]
	// MaxLength is the maximum number of frames to keep in the trackfile.
	// At least 3 are needed to distinguish linear tracks and curved tracks. More than 5 is probably excessive.
	MaxLength int
}

// Frame describes a contact's position and velocity at a point in time.
type Frame struct {
	// Timestamp of the observation.
	Timestamp time.Time
	// Point is the contact's 2D position.
	Point orb.Point
	// Altitude above sea level.
	Altitude unit.Length
	// Heading is the direction the contact is moving. This is not necessarily the direction the nose is poining.
	Heading unit.Angle
}

func NewTrackfile(a Aircraft) *Trackfile {
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

func (t *Trackfile) Update(f Frame) {
	t.Track.PushFront(f)
	for t.Track.Len() > t.MaxLength {
		t.Track.PopBack()
	}
}

func (t *Trackfile) Bullseye(bullseye orb.Point) brevity.Bullseye {
	latest := t.Track.Front()
	bearing := unit.Angle(geo.Bearing(bullseye, latest.Point)) * unit.Degree
	distance := unit.Length(geo.Distance(bullseye, latest.Point)) * unit.Meter
	return *brevity.NewBullseye(bearing, distance)
}

func (t *Trackfile) LastKnown() Frame {
	if t.Track.Len() == 0 {
		return Frame{
			Timestamp: time.Now().Add(-time.Hour),
			Point:     orb.Point{},
			Altitude:  0,
			Heading:   0,
		}
	}
	return t.Track.Front()
}

func (t *Trackfile) Course() unit.Angle {
	if t.Track.Len() < 2 {
		return unit.Angle(t.LastKnown().Heading) * unit.Degree
	}

	latest := t.Track.Front()
	previous := t.Track.At(1)

	course := geo.Bearing(previous.Point, latest.Point)
	if course < 0 {
		course += 360
	}
	course = math.Mod(course, 360)
	if course < 1 {
		course = 360.0
	}

	return unit.Angle(course) * unit.Degree
}

func (t *Trackfile) Direction() brevity.Track {
	if t.Track.Len() < 1 {
		return brevity.UnknownDirection
	}
	course := t.Course()
	return brevity.TrackFromBearing(course)
}

func (t *Trackfile) Speed() unit.Speed {
	if t.Track.Len() < 2 {
		return 0
	}

	latest := t.Track.Front()
	previous := t.Track.At(1)

	timeDelta := latest.Timestamp.Sub(previous.Timestamp)

	groundDistance := unit.Length(geo.Distance(latest.Point, previous.Point)) * unit.Meter
	groundSpeed := unit.Speed(
		groundDistance.Meters()/
			timeDelta.Seconds(),
	) * unit.MetersPerSecond

	distanceVertical := latest.Altitude - previous.Altitude
	verticalSpeed := unit.Speed(
		distanceVertical.Meters()/
			timeDelta.Seconds(),
	) * unit.MetersPerSecond

	trueSpeed := unit.Speed(
		math.Sqrt(
			math.Pow(groundSpeed.MetersPerSecond(), 2)+
				math.Pow(verticalSpeed.MetersPerSecond(), 2),
		),
	) * unit.MetersPerSecond

	if groundSpeed > trueSpeed {
		return groundSpeed
	}
	return trueSpeed
}
