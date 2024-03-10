package brevity

import "github.com/martinlindhe/unit"

// Bullseye is a magnetic bearing and distance from a reference point called the BULLSEYE.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection a
type Bullseye interface {
	// Bearing from the BULLSEYE to the contact, rounded to the nearest degree.
	Bearing() unit.Angle
	// Distance from the BULLSEYE to the contact, rounded to the nearest nautical mile.
	Distance() unit.Length
}
