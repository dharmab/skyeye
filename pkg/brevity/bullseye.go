package brevity

import (
	"math"

	"github.com/martinlindhe/unit"
)

// Bullseye is a magnetic bearing and distance from a reference point called the BULLSEYE.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection a
type Bullseye interface {
	// Bearing from the BULLSEYE to the contact, rounded to the nearest degree.
	Bearing() unit.Angle
	// Distance from the BULLSEYE to the contact, rounded to the nearest nautical mile.
	Distance() unit.Length
}

var _ Bullseye = &bullseye{}

type bullseye struct {
	bearing  unit.Angle
	distance unit.Length
}

func NewBullseye(bearing unit.Angle, distance unit.Length) Bullseye {
	return &bullseye{
		bearing:  bearing,
		distance: distance,
	}
}

func (b *bullseye) Bearing() unit.Angle {
	a := math.Round(b.bearing.Degrees())
	return unit.Angle(a) * unit.Degree
}

func (b *bullseye) Distance() unit.Length {
	d := math.Round(b.distance.NauticalMiles())
	return unit.Length(d) * unit.NauticalMile
}
