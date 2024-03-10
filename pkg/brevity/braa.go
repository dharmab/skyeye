package brevity

import "github.com/martinlindhe/unit"

// BRAA provides target bearing, range, altitude and aspect relative to a specified friendly aircraft.
// Reference: ATP 3-52.4 Chapter IV section 4 subsection b
type BRAA interface {
	// Bearing is the heading from the fighter to the contact, rounded to the nearest degree.
	Bearing() unit.Angle
	// Range is the distance from the fighter to the contact, rounded to the nearest nautical mile.
	Range() unit.Length
	// Altitude of the contact above sea level, rounded to the nearest thousands of feet.
	Altitude() unit.Length
	// Aspect of the contact.
	Aspect() Aspect
}
