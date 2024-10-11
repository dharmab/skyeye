package brevity

import "fmt"

// SnaplockRequest is an abbreviated form of DECLARE used to quickly gain infomation on a contact inside THREAT range with BEAM or hotter aspect.
// Aspect is implied to be Beam or greater.
// Reference ATP 3-52.4 Chapter V section 20.
type SnaplockRequest struct {
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign string
	// BRA is the location of the contact.
	BRA BRA
}

func (r SnaplockRequest) String() string {
	return fmt.Sprintf("SNAPLOCK for %s: bra %s", r.Callsign, r.BRA)
}

// SnaplockResponse is a response to a SNAPLOCK call.
// Reference ATP 3-52.4 Chapter V section 20.
type SnaplockResponse struct {
	// Callsign of the friendly aircraft requesting the SNAPLOCK.
	Callsign string
	// Declaration of the contact.
	Declaration Declaration
	// Group that was identified. If Declaration is Unable or Furball, this may be nil.
	Group Group
}
