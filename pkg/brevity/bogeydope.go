package brevity

// BogeyDopeRequest is a request for a BOGEY DOPE.
// Reference: ATP 3-52.4 Chapter V section 11
type BogeyDopeRequest interface {
	// Callsign of the friendly aircraft  requesting the BOGEY DOPE.
	Callsign() string
}

type BogeyDopeResponse interface {
	// Callsign of the friendly aircraft requesting the BOGEY DOPE.
	Callsign() string
	// Group which is closest to the fighter.
	Group() Group
}
