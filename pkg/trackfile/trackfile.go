package trackfile

import (
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
)

type Aircraft struct {
	// UnitID is the in-game ID.
	UnitID uint32
	// Name is a unique string for each aircraft.
	// If this is a player's aircraft, it will be the player's in-game name.
	Name string
	// Coalition the aircraft belongs to.
	Coalition types.Coalition
	// The string for the aircraft type in DCS. This is sometimes a weird string like FA18C_hornet, A-10C_2, or F-15ESE.
	// Use github.com/dharmab/skyeye/pkg/encyclopedia to look up the real world type.
	EditorType string
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
	// Speed is ground speed.
	Speed unit.Speed
}

func NewTrackfile(a Aircraft) *Trackfile {
	return &Trackfile{
		Contact:   a,
		Track:     *deque.New[Frame](),
		MaxLength: 4,
	}
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
	distance := unit.Length(planar.Distance(bullseye, latest.Point)) * unit.Meter
	return *brevity.NewBullseye(bearing, distance)
}

func (t *Trackfile) LastKnown() Frame {
	return t.Track.Front()
}
