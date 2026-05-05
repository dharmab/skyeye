package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// BRAA provides target bearing, range, altitude and aspect relative to a specified friendly aircraft.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection b.
type BRAA struct {
	BRA
	// aspect of the contact.
	aspect Aspect
}

// BRA is an abbreviated form of BRAA without aspect.
type BRA struct {
	Vector
	stacks []Stack
}

// NewBRA creates a new [BRA].
func NewBRA(b bearings.Bearing, r unit.Length, a ...unit.Length) *BRA {
	if !b.IsMagnetic() {
		log.Warn().Stringer("bearing", b).Msg("bearing provided to NewBRA should be magnetic")
	}
	return &BRA{
		Vector: Vector{
			bearing:  b,
			distance: r,
		},
		stacks: Stacks(a...),
	}
}

// Altitude returns the highest altitude of the contact above sea level, rounded to the nearest thousands of feet.
func (b *BRA) Altitude() unit.Length {
	if len(b.stacks) == 0 {
		return 0
	}
	return spatial.NormalizeAltitude(b.stacks[0].Altitude)
}

// Stacks returns the altitude STACKS of the contact.
func (b *BRA) Stacks() []Stack {
	return b.stacks
}

func (b *BRA) String() string {
	s := fmt.Sprintf("BRA %s/%.0f %.0f", b.Bearing(), b.Range().NauticalMiles(), b.Altitude().Feet())
	if len(b.Stacks()) > 1 {
		s += fmt.Sprintf(" (%v)", b.Stacks())
	}
	return s
}

// NewBRAA creates a new [BRAA].
func NewBRAA(b bearings.Bearing, r unit.Length, altitudes []unit.Length, aspect Aspect) *BRAA {
	if !b.IsMagnetic() {
		log.Warn().Stringer("bearing", b).Msg("bearing provided to NewBRAA should be magnetic")
	}
	return &BRAA{
		BRA: BRA{
			Vector: Vector{
				bearing:  b,
				distance: r,
			},
			stacks: Stacks(altitudes...),
		},
		aspect: aspect,
	}
}

// Aspect returns the aspect of the contact.
func (b *BRAA) Aspect() Aspect {
	return b.aspect
}

func (b *BRAA) String() string {
	return fmt.Sprintf("%s %s", &b.BRA, b.Aspect())
}
