// package bearings conttains functions for working with absolute and magnetic bearings.
package bearings

import (
	"github.com/martinlindhe/unit"
)

// Bearing is a compass bearing that can be converted to true or magnetic bearing. Use either True or Magnetic immediately before use to assert which kind is used.
type Bearing interface {
	// Value returns the compass heading, normalized in range (0, 360].
	Value() unit.Angle
	// Rounded returns the compass heading, normalized in range (0, 360] rounded to the nearest degree.
	Rounded() unit.Angle
	// Degrees returns the compass heading in degrees, normalized in range (0, 360].
	Degrees() float64
	// RoundedDegrees returns the compass heading in degrees, normalized in range (0, 360] rounded to the nearest degree.
	RoundedDegrees() float64
	// True bearing conversion computed with the given declination.
	True(declination unit.Angle) Bearing
	// Magnetic bearing conversion computed with the given declination.
	Magnetic(declination unit.Angle) Bearing
	// IsTrue returns true if the bearing is a true bearing.
	IsTrue() bool
	// IsMagnetic returns true if the bearing is a magnetic bearing.
	IsMagnetic() bool
	// String returns the value as a three-digit string.
	String() string
}
