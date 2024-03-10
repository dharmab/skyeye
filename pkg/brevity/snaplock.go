package brevity

// SnaplockRequest is an abbreviated form of DECLARE used to quickly gain infomation on a contact inside THREAT range with BEAM or hotter aspect.
// Reference ATP 3-52.4 Chapter V section 20
type SnaplockRequest interface {
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign() string
	// BRAA is the location of the contact.
	BRAA() BRAA
}

// SnaplockResponse is a response to a SNAPLOCK call.
// Reference ATP 3-52.4 Chapter V section 20
type SnaplockResponse interface {
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign() string
	// Status is true if the SNAPLOCK was correlated to a group, otherwise false.
	Status() bool
	// Group that was identified. If Status is false, this may be nil.
	Group() Group
}
