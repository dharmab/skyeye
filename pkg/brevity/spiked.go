package brevity

import "github.com/martinlindhe/unit"

// SpikedRequest is a request to correlate a radar spike within ±30 degrees.
// Reference: ATP 3-52.4 Chapter V section 13
type SpikedRequest struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// Bearing to the radar spike.
	Bearing unit.Angle
}

// SpikedResponse reports any contacts within ±30 degrees of a reported radar spike.
// Reference: ATP 3-52.4 Chapter V section 13
type SpikedResponse struct {
	// Callsign of the friendly aircraft calling SPIKED.
	Callsign string
	// True if the spike was correlated to a contact. False otherwise.
	Status bool
	// Range to the correlated contact. If Status is false, this may be nil.
	Range *unit.Length
	// Altitude of the correlated contact. If Status is false, this may be nil.
	Altitude *unit.Length
	// Aspect of the correlated contact. If Status is false, this may be UnknownAspect.
	Aspect Aspect
	// Declaration of the correlated contact. If Status is false, this may be Clean.
	Declaration Declaration
	// Number of contacts in the correlated group. If Status is false, this may be zero.
	Contacts int
	// Reported spike bearing. This is used if the response did not correlate to a group.
	Bearing unit.Angle
}
