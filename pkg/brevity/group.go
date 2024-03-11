package brevity

import "github.com/martinlindhe/unit"

// Group describes any number of air contacts within 3 nautical miles in azimuth and range of each other.
// Groups are distinguished by either a unique name or a location. This implementation only uses location.
// Location may be either BRAA or Bullseye, altitude and track. Bullseye is preferred except for:
//
// - BOGEY DOPE and SNAPLOCK responses
//
// - THREAT calls that pertain to a single aircraft
//
// Reference: ATP 3-52.4 Chapter IV section 2
type Group interface {
	// Contacts is the number of contacts in the group.
	Contacts() int
	// Bullseye is the location of the group. This may be nil for BOGEY DOPE, SNAPLOCK, and THREAT calls.
	Bullseye() Bullseye
	// Altitude is the group's altitude above sea level. This may be nil for BOGEY DOPE, SNAPLOCK, and THREAT calls.
	Altitude() unit.Length
	// Track is the group's track direction. This may be nil for BOGEY DOPE, SNAPLOCK, and THREAT calls.
	Track() unit.Angle
	// BRAA is an alternate format for the group's location. This is nil except for BOGEY DOPE, SNAPLOCK, and some THREAT calls.
	BRAA() BRAA
	// Declaration of the group's friend or foe status.
	Declaration() Declaration
	// Heavy is true if the group contacts 3 or more contacts.
	Heavy() bool
	// Type is the NATO reporting name of the group's aircraft platform (for Soviet/Russian/Chinese aircraft) or
	Type() string
	// High is true if the aircraft altitude is above 40,000 feet.
	High() bool
	// Fast is true if the group's speed is 600-900kts ground speed or 1.0-1.5 Mach.
	Fast() bool
	// VeryFast is true is the group's speed is above 900kts ground speed or 1.5 Mach.
	VeryFast() bool
}
