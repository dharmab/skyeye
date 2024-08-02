package bearings

import (
	"math"

	"github.com/martinlindhe/unit"
)

type True struct {
	θ unit.Angle
}

var _ Bearing = True{}

func NewTrueBearing(value unit.Angle) True {
	return True{θ: Normalize(value)}
}

func (b True) Value() unit.Angle {
	return Normalize(b.θ)
}

func (b True) Rounded() unit.Angle {
	return unit.Angle(b.RoundedDegrees()) * unit.Degree
}

func (b True) Degrees() float64 {
	return b.Value().Degrees()
}

func (b True) RoundedDegrees() float64 {
	return math.Round(b.Degrees())
}

func (b True) True(declination unit.Angle) Bearing {
	return b
}

func (b True) Magnetic(declination unit.Angle) Bearing {
	return NewMagneticBearing(b.Value() - declination)
}

func (b True) IsTrue() bool {
	return true
}

func (b True) IsMagnetic() bool {
	return false
}

func (b True) String() string {
	return toString(b)
}
