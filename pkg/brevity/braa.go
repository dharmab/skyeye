package brevity

import (
	"math"

	"github.com/martinlindhe/unit"
)

// BRAA provides target bearing, range, altitude and aspect relative to a specified friendly aircraft.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection b
type BRAA interface {
	BRA
	// Aspect of the contact.
	Aspect() Aspect
}

// BRA is an abbreviated form of BRAA without aspect.
type BRA interface {
	// Bearing is the heading from the fighter to the contact, rounded to the nearest degree.
	Bearing() unit.Angle
	// Range is the distance from the fighter to the contact, rounded to the nearest nautical mile.
	Range() unit.Length
	// Altitude of the contact above sea level, rounded to the nearest thousands of feet.
	Altitude() unit.Length
}

type bra struct {
	bearingDegrees      int
	rangeNM             int
	altitudeThousandsFt int
}

func NewBRA(b unit.Angle, r unit.Length, a unit.Length) BRA {
	return &bra{
		bearingDegrees:      int(math.Round(b.Degrees())),
		rangeNM:             int(math.Round(r.NauticalMiles())),
		altitudeThousandsFt: int(math.Round(a.Feet() / 1000)),
	}
}

// Bearing implements [BRA.Bearing].
func (b *bra) Bearing() unit.Angle {
	return unit.Angle(b.bearingDegrees) * unit.Degree
}

// Range implements [BRA.Range].
func (b *bra) Range() unit.Length {
	return unit.Length(b.rangeNM) * unit.NauticalMile
}

// Altitude implements [BRA.Altitude].
func (b *bra) Altitude() unit.Length {
	return unit.Length(b.altitudeThousandsFt*1000) * unit.Foot
}

type braa struct {
	bra    BRA
	aspect Aspect
}

func NewBRAA(b unit.Angle, r unit.Length, a unit.Length, aspect Aspect) BRAA {
	return &braa{
		bra:    NewBRA(b, r, a),
		aspect: aspect,
	}
}

// Bearing implements [BRA.Bearing].
func (b *braa) Bearing() unit.Angle {
	return b.bra.Bearing()
}

// Range implements [BRA.Range].
func (b *braa) Range() unit.Length {
	return b.bra.Range()
}

// Altitude implements [BRA.Altitude].
func (b *braa) Altitude() unit.Length {
	return b.bra.Altitude()
}

// Aspect implements [BRAA.Aspect].
func (b *braa) Aspect() Aspect {
	return b.aspect
}
