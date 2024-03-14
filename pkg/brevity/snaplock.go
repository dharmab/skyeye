package brevity

// SnaplockRequest is an abbreviated form of DECLARE used to quickly gain infomation on a contact inside THREAT range with BEAM or hotter aspect.
// Aspect is implied to be Beam or greater.
// Reference ATP 3-52.4 Chapter V section 20
type SnaplockRequest interface {
	Snaplock()
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign() string
	// BRA is the location of the contact.
	BRA() BRA
}

// SnaplockResponse is a response to a SNAPLOCK call.
// Reference ATP 3-52.4 Chapter V section 20
type SnaplockResponse interface {
	Snaplock()
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign() string
	// Status is true if the SNAPLOCK was correlated to a group, otherwise false.
	Status() bool
	// Group that was identified. If Status is false, this may be nil.
	Group() Group
}
