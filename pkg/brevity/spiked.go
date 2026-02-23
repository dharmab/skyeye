package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
)

// SpikedRequest is a request to correlate a radar spike within ±30 degrees.
// Reference: ATP 3-52.4 Chapter V section 13.
type SpikedRequest struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// Bearing to the radar spike.
	Bearing bearings.Bearing
}

func (r SpikedRequest) String() string {
	return fmt.Sprintf("SPIKED for %s: bearing %s", r.Callsign, r.Bearing)
}

// SpikedResponse reports any contacts within ±30 degrees of a reported radar spike.
// Reference: ATP 3-52.4 Chapter V section 13.
type SpikedResponse struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// Reported spike bearing. This is used if the response did not correlate to a group.
	Bearing bearings.Bearing
	// True if the spike was correlated to a contact. False otherwise.
	Status bool
	// Correleted contact group. If Status is false, this may be nil.
	Group Group
}
