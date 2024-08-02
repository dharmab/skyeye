// package bearings conttains functions for working with absolute and magnetic bearings.
package bearings

import (
	"math"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
)

type Kind int

const (
	True Kind = iota
	Magnetic
)

type Bearing struct {
	value unit.Angle
	kind  Kind
}

func New(kind Kind, value unit.Angle) Bearing {
	return Bearing{value: Normalize(value), kind: kind}
}

func (b Bearing) Value() unit.Angle {
	return b.value
}

func (b Bearing) Kind() Kind {
	return b.kind
}

// Normalize returns the normalized angle in the range (0, 360] degrees, rounded to the nearest degree.
func Normalize(a unit.Angle) unit.Angle {
	θ := float64(a.Degrees())
	for θ < 0 {
		θ += 360
	}
	θ = math.Mod(θ, 360)
	if θ == 0 {
		θ = 360
	}
	return unit.Angle(θ) * unit.Degree
}

func Variation(p orb.Point, t time.Time) unit.Angle {
	// TODO compute from model
	return unit.Angle(0) * unit.Degree
}

func TrueToMagnetic(tru unit.Angle, variation unit.Angle) unit.Angle {
	return Normalize(tru + variation)
}

func MagneticToTrue(magnetic unit.Angle, variation unit.Angle) unit.Angle {
	return Normalize(magnetic - variation)
}
