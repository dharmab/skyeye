package bearings

import (
	"math"

	"github.com/martinlindhe/unit"
)

type Magnetic struct {
	θ unit.Angle
}

func NewMagneticBearing(value unit.Angle) *Magnetic {
	return &Magnetic{θ: Normalize(value)}
}

var _ Bearing = &Magnetic{}

func (b *Magnetic) Value() unit.Angle {
	return Normalize(b.θ)
}

func (b *Magnetic) Rounded() unit.Angle {
	return unit.Angle(b.RoundedDegrees()) * unit.Degree
}

func (b *Magnetic) Degrees() float64 {
	return b.Value().Degrees()
}

func (b *Magnetic) RoundedDegrees() float64 {
	return math.Round(b.Degrees())
}

func (b *Magnetic) True(declination unit.Angle) Bearing {
	return NewTrueBearing(b.Value() + declination)
}

func (b *Magnetic) Magnetic(declination unit.Angle) Bearing {
	return b
}

func (b *Magnetic) IsTrue() bool {
	return false
}

func (b *Magnetic) IsMagnetic() bool {
	return true
}

func (b *Magnetic) String() string {
	return toString(b)
}
