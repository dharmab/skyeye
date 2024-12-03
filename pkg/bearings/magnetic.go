package bearings

import (
	"math"

	"github.com/martinlindhe/unit"
)

// Magnetic is a magnetic bearing, pointing at the magnetic north pole.
type Magnetic struct {
	θ unit.Angle
}

var _ Bearing = &Magnetic{}

// NewMagneticBearing creates a new magnetic bearing from the given value.
func NewMagneticBearing(value unit.Angle) *Magnetic {
	return &Magnetic{θ: normalize(value)}
}

// Value returns the magnetic bearing value.
func (b *Magnetic) Value() unit.Angle {
	return normalize(b.θ)
}

// Rounded returns the magnetic bearing rounded to the nearest degree.
func (b *Magnetic) Rounded() unit.Angle {
	return unit.Angle(b.RoundedDegrees()) * unit.Degree
}

// Degrees returns the magnetic bearing in degrees.
func (b *Magnetic) Degrees() float64 {
	return b.Value().Degrees()
}

// RoundedDegrees returns the magnetic bearing in degrees, rounded to the nearest degree.
func (b *Magnetic) RoundedDegrees() float64 {
	return math.Round(b.Degrees())
}

// True converts this magnetic bearing to a true bearing by removing the given declination.
func (b *Magnetic) True(declination unit.Angle) Bearing {
	return NewTrueBearing(b.Value() + declination)
}

// Magnetic returns this magnetic bearing.
func (b *Magnetic) Magnetic(_ unit.Angle) Bearing {
	return b
}

// Reciprocal returns a magnetic reciprocal.
func (b *Magnetic) Reciprocal() Bearing {
	return NewMagneticBearing(b.Value() + 180*unit.Degree)
}

// IsTrue returns false for a magnetic bearing.
func (*Magnetic) IsTrue() bool {
	return false
}

// IsMagnetic returns true for a magnetic bearing.
func (*Magnetic) IsMagnetic() bool {
	return true
}

// String implements [Bearing.String].
func (b *Magnetic) String() string {
	return toString(b)
}
