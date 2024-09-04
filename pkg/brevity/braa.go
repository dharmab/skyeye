package brevity

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// BRAA provides target bearing, range, altitude and aspect relative to a specified friendly aircraft.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection b.
type BRAA interface {
	BRA
	// Aspect of the contact.
	Aspect() Aspect
}

// BRA is an abbreviated form of BRAA without aspect.
type BRA interface {
	// Bearing is the heading from the fighter to the contact, rounded to the nearest degree.
	Bearing() bearings.Bearing
	// Range is the distance from the fighter to the contact, rounded to the nearest nautical mile.
	Range() unit.Length
	// Altitude of the contact above sea level, rounded to the nearest thousands of feet.
	Altitude() unit.Length
	// Altitude STACKS of the contact above sea level, rounded to the nearest thousands of feet.
	Stacks() []Stack
}

type bra struct {
	bearing bearings.Bearing
	_range  unit.Length
	stacks  []Stack
}

func NewBRA(b bearings.Bearing, r unit.Length, a ...unit.Length) BRA {
	if !b.IsMagnetic() {
		log.Warn().Stringer("bearing", b).Msg("bearing provided to NewBRA should be magnetic")
	}
	return &bra{
		bearing: b,
		_range:  r,
		stacks:  Stacks(a...),
	}
}

// Bearing implements [BRA.Bearing].
func (b *bra) Bearing() bearings.Bearing {
	return b.bearing
}

// Range implements [BRA.Range].
func (b *bra) Range() unit.Length {
	return unit.Length(math.Round(b._range.NauticalMiles())) * unit.NauticalMile
}

// Altitude implements [BRA.Altitude].
func (b *bra) Altitude() unit.Length {
	if len(b.stacks) == 0 {
		return 0
	}
	return spatial.NormalizeAltitude(b.stacks[0].Altitude)
}

// Stacks implements [BRA.Stacks].
func (b *bra) Stacks() []Stack {
	return b.stacks
}

type braa struct {
	bra    BRA
	aspect Aspect
}

func NewBRAA(b bearings.Bearing, r unit.Length, a []unit.Length, aspect Aspect) BRAA {
	if !b.IsMagnetic() {
		log.Warn().Stringer("bearing", b).Msg("bearing provided to NewBRAA should be magnetic")
	}
	return &braa{
		bra:    NewBRA(b, r, a...),
		aspect: aspect,
	}
}

// Bearing implements [BRA.Bearing].
func (b *braa) Bearing() bearings.Bearing {
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

// Stacks implements [BRA.Stacks].
func (b *braa) Stacks() []Stack {
	return b.bra.Stacks()
}

// Aspect implements [BRAA.Aspect].
func (b *braa) Aspect() Aspect {
	return b.aspect
}
