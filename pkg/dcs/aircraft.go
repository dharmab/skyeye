package dcs

import (
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	measure "github.com/martinlindhe/unit"
)

type Aircraft interface {
	// Timestamp returns the real-world clock time of the observation.
	Timestamp() time.Time
	// UnitID returns the unit's in-game ID.
	UnitID() uint32
	// Name returns the player name if this is a player-controlled unit, or the unit name defined in the mission editor.
	Name() string
	// Coalition returns the unit's coalition.
	Coalition() common.Coalition
	// Platform is the in-game unit type. (Avoided use of word "Type" which has another meaning in Go.)
	Platform() string
	// Latitude in decimal degrees.
	Latitude() measure.Angle
	// Longitude in decimal degrees.
	Longitude() measure.Angle
	// Altitude in meters above sea level.
	Altitude() measure.Length
	// Heading is the aircraft's true course.
	// This is the direction the aircraft is moving, not necessarily the direction the nose is pointing.
	Heading() measure.Angle
	// Speed is true ground speed.
	// This is a two dimensional measure. An aircraft in a vertical climb or dive will have a speed near zero.
	Speed() measure.Speed
}

type aircraft struct {
	timestamp time.Time
	unitID    uint32
	name      string
	coalition common.Coalition
	platform  string
	latitude  measure.Angle
	longitude measure.Angle
	altitude  measure.Length
	heading   measure.Angle
	speed     measure.Speed
}

var _ Aircraft = &aircraft{}

func (a *aircraft) Timestamp() time.Time {
	return a.timestamp
}

func (a *aircraft) UnitID() uint32 {
	return a.unitID
}

func (a *aircraft) Name() string {
	return a.name
}

func (a *aircraft) Coalition() common.Coalition {
	return a.coalition
}

func (a *aircraft) Platform() string {
	return a.platform
}

func (a *aircraft) Latitude() measure.Angle {
	return a.latitude
}

func (a *aircraft) Longitude() measure.Angle {
	return a.longitude
}

func (a *aircraft) Altitude() measure.Length {
	return a.altitude
}

func (a *aircraft) Heading() measure.Angle {
	return a.heading
}

func (a *aircraft) Speed() measure.Speed {
	return a.speed
}
