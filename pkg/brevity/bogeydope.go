package brevity

import (
	"fmt"
	"strconv"
)

type ContactCategory int

const (
	Aircraft ContactCategory = iota
	FixedWing
	RotaryWing
)

func (c ContactCategory) String() string {
	switch c {
	case Aircraft:
		return "Aircraft"
	case FixedWing:
		return "Fixed Wing"
	case RotaryWing:
		return "Rotary Wing"
	}
	return strconv.Itoa(int(c))
}

// BogeyDopeRequest is a request for a BOGEY DOPE.
// Reference: ATP 3-52.4 Chapter V section 11.
type BogeyDopeRequest struct {
	// Callsign of the friendly aircraft requesting the BOGEY DOPE.
	Callsign string
	// Filter for the type of aircraft to include in the BOGEY DOPE.
	Filter ContactCategory
}

func (r BogeyDopeRequest) String() string {
	return fmt.Sprintf("BOGEY DOPE for %s: filter %v", r.Callsign, r.Filter)
}

type BogeyDopeResponse struct {
	// Callsign of the friendly aircraft requesting the BOGEY DOPE.
	Callsign string
	// Group which is closest to the fighter. If there are no eligible groups, this may be nil.
	Group Group
}
