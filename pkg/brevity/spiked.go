package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
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
//
// Deprecated: Use SpikedResponseV2 instead.
type SpikedResponse struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// True if the spike was correlated to a contact. False otherwise.
	Status bool
	// Range to the correlated contact. If Status is false, this may be 0.
	Range unit.Length
	// Altitude of the correlated contact. If Status is false, this may be 0.
	Altitude unit.Length
	// Aspect of the correlated contact. If Status is false, this may be UnknownAspect.
	Aspect Aspect
	// Track of the correlated contact. If Status is false, this may be UnknownDirection.
	Track Track
	// Declaration of the correlated contact. If Status is false, this may be Clean.
	Declaration Declaration
	// Number of contacts in the correlated group. If Status is false, this may be zero.
	Contacts int
	// Reported spike bearing. This is used if the response did not correlate to a group.
	Bearing bearings.Bearing
}

// SpikedResponseV2 reports any contacts within ±30 degrees of a reported radar spike.
// Reference: ATP 3-52.4 Chapter V section 13.
type SpikedResponseV2 struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// Reported spike bearing. This is used if the response did not correlate to a group.
	Bearing bearings.Bearing
	// True if the spike was correlated to a contact. False otherwise.
	Status bool
	// Correleted contact group. If Status is false, this may be nil.
	Group Group
}
