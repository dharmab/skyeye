package brevity

import (
	"fmt"
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// Bullseye is a magnetic bearing and distance from a reference point called the BULLSEYE.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection a.
type Bullseye struct {
	bearing  bearings.Bearing
	distance unit.Length
}

func NewBullseye(bearing bearings.Bearing, distance unit.Length) *Bullseye {
	if !bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to NewBullseye should be magnetic")
	}
	return &Bullseye{
		bearing:  bearing,
		distance: distance,
	}
}

// Bearing from the BULLSEYE to the contact, rounded to the nearest degree.
func (b *Bullseye) Bearing() bearings.Bearing {
	return b.bearing
}

// Distance from the BULLSEYE to the contact, rounded to the nearest nautical mile.
func (b *Bullseye) Distance() unit.Length {
	return unit.Length(math.Round(b.distance.NauticalMiles())) * unit.NauticalMile
}

func (b Bullseye) String() string {
	return fmt.Sprintf("%s/%.0f", b.bearing, b.distance.NauticalMiles())
}
