package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
)

// Strobe is a request to correlate an electromagnetic attack within ±30 degrees.
// Reference: ATP 3-52.4 Chapter V section 12.
type StrobeRequest struct {
	// Callsign of the friendly aircraft calling STROBE..
	Callsign string
	// Bearing to the electromagnetic attack.
	Bearing bearings.Bearing
}

func (r StrobeRequest) String() string {
	return fmt.Sprintf("STROBE for %s: bearing %s", r.Callsign, r.Bearing)
}

// StrobeResponse reports any contacts within ±30 degrees of a reported electromagnetic attack.
// Reference: ATP 3-52.4 Chapter V section 12.
type StrobeResponse struct {
	// Callsign of the friendly aircraft calling STROBE..
	Callsign string
	// Reported attack bearing. This is used if the response did not correlate to a group.
	Bearing bearings.Bearing
	// True if the attack was correlated to a contact. False otherwise.
	Status bool
	// Correleted contact group. If Status is false, this may be nil.
	Group Group
}
