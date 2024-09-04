package bearings

import (
	"math"

	"github.com/martinlindhe/unit"
)

// True is a bearing pointing to geographic north.
type True struct {
	θ unit.Angle
}

var _ Bearing = True{}

// NewTrueBearing creates a new bearing from the given value.
func NewTrueBearing(value unit.Angle) True {
	return True{θ: normalize(value)}
}

// Value returns the bearing value.
func (b True) Value() unit.Angle {
	return normalize(b.θ)
}

// Rounded returns the bearing rounded to the nearest degree.
func (b True) Rounded() unit.Angle {
	return unit.Angle(b.RoundedDegrees()) * unit.Degree
}

// Degrees returns the bearing in degrees.
func (b True) Degrees() float64 {
	return b.Value().Degrees()
}

// RoundedDegrees returns the bearing in degrees, rounded to the nearest degree.
func (b True) RoundedDegrees() float64 {
	return math.Round(b.Degrees())
}

// True returns this bearing.
func (b True) True(declination unit.Angle) Bearing {
	return b
}

// Magnetic converts this true bearing to a magnetic bearing by subtracting the given declination.
func (b True) Magnetic(declination unit.Angle) Bearing {
	return NewMagneticBearing(b.Value() - declination)
}

// Reciprocal returns a reciprocal true bearing.
func (b True) Reciprocal() Bearing {
	return NewTrueBearing(b.Value() + 180*unit.Degree)
}

// IsTrue returns true for a true bearing.
func (b True) IsTrue() bool {
	return true
}

// IsMagnetic returns false for a true bearing.
func (b True) IsMagnetic() bool {
	return false
}

// String implements [Bearing.String].
func (b True) String() string {
	return toString(b)
}
