package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// Bullseye is a magnetic bearing and distance from a reference point called the BULLSEYE.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection a.
type Bullseye interface {
	Bearing() bearings.Bearing
	Distance() unit.Length
}

type bullseye struct {
	vector
}

// NewBullseye creates a new [Bullseye].
func NewBullseye(bearing bearings.Bearing, distance unit.Length) Bullseye {
	if !bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to NewBullseye should be magnetic")
	}
	return &bullseye{
		vector: vector{
			bearing:  bearing,
			distance: distance,
		},
	}
}

// Bearing from the BULLSEYE to the contact, rounded to the nearest degree.
func (b *bullseye) Bearing() bearings.Bearing {
	return b.vector.Bearing()
}

// Distance from the BULLSEYE to the contact, rounded to the nearest nautical mile.
func (b *bullseye) Distance() unit.Length {
	return b.Range()
}

func (b *bullseye) String() string {
	return fmt.Sprintf("%s/%.0f", b.Bearing(), b.Distance().NauticalMiles())
}
