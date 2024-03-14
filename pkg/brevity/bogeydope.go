package brevity

type BogeyFilter int

const (
	Everything BogeyFilter = iota
	Airplanes
	Helicopters
)

// BogeyDopeRequest is a request for a BOGEY DOPE.
// Reference: ATP 3-52.4 Chapter V section 11
type BogeyDopeRequest interface {
	BogeyDope()
	// Callsign of the friendly aircraft requesting the BOGEY DOPE.
	Callsign() string
	// Filter for the type of aircraft to include in the BOGEY DOPE.
	Filter() BogeyFilter
}

type BogeyDopeResponse interface {
	BogeyDope()
	// Callsign of the friendly aircraft requesting the BOGEY DOPE.
	Callsign() string
	// Group which is closest to the fighter.
	Group() Group
}
