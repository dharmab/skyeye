package brevity

import (
	"math"

	"github.com/martinlindhe/unit"
)

// Bullseye is a magnetic bearing and distance from a reference point called the BULLSEYE.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection a
type Bullseye struct {
	bearingDegrees int
	distanceNM     int
}

func NewBullseye(bearing unit.Angle, distance unit.Length) *Bullseye {
	return &Bullseye{
		bearingDegrees: int(math.Round(bearing.Degrees())),
		distanceNM:     int(math.Round(distance.NauticalMiles())),
	}
}

// Bearing from the BULLSEYE to the contact, rounded to the nearest degree.
func (b *Bullseye) Bearing() unit.Angle {
	return unit.Angle(b.bearingDegrees) * unit.Degree
}

// Distance from the BULLSEYE to the contact, rounded to the nearest nautical mile.
func (b *Bullseye) Distance() unit.Length {
	return unit.Length(b.distanceNM) * unit.NauticalMile
}
