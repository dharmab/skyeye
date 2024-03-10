package brevity

import "github.com/martinlindhe/unit"

// Reference ATP 3-52.4 Chapter V section 6
type Declaration int

const (
	// Bogey indicates the contact's' whose identity is unknown.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Bogey Declaration = iota
	// Friendly indicates the contact is a positively identified friendly.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Friendly
	// Neutral indicates the contact is a positively identified aircraft whose characteristics, behavior, origin or nationality indicate it is neither supporting nor opposing friendly forces.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Neutral
	// Bandit indicates the contact is a positively idenfieid enemy in accordance with theater identification criteria. It does not imply direction or authority to engage.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Bandit
	// Hostile indicates the contact is a identified as an enemy upon which clearance to fire is authorized in accordance with theater rules of engagement.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Hostile
	// DeclarationFurball indicates that non-friendly and friendly aircraft are inside of 5 nauctical miles of each other.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Furball
	// Unable indications that the responder is unable to provide a declaration as requested.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Unable
	// Clean indicates there is no sensor information on the contact.
	// Reference: ATP 1-02.1 Chapter I Table 2
	Clean
)

// DeclareRequest is a DECLARE call.
// Reference: ATP 3-52.4 Chapter V section 6
type DeclareRequest interface {
	// Callsign of the friendly aircraft requesting DECLARE.
	Callsign() string
	// Location of the contact.
	Location() Bullseye
	// Altitude of the contact above sea level, rounded to the nearest thousands of feet.
	Altitude() unit.Length
	// Track direction. Optional, used to discriminate between multiple contacts at the same location.
	Track() CardinalDirection
}

// DeclareResponse is a response to a DECLARE call.
// Reference: ATP 3-52.4 Chapter V section 6
type DeclareResponse interface {
	// Callsign of the friendly aircraft requesting DECLARE.
	Callsign() string
	// Declaration of the contact.
	Declaration() Declaration
	// Group that was identified, if a specific one was identifiable.
	// This may be nil if Declaration is Furball, Unable, or Clean.
	Group() Group
}
