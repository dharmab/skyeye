package brevity

// AlphaCheckRequest is a request for an ALPHA CHECK.
// An ALPHA CHECK is a request for the friendly aircraft's position.
// It is used by aircrews to check their position equipment, especially for aircraft without GPS.
// Reference: ATP 3-52.4 Chapter II section 4
type AlphaCheckRequest struct {
	// Callsign of the friendly aircraft requesting the ALPHA CHECK.
	Callsign string
}

// AlphaCheckResponse is a response to an ALPHA CHECK.
type AlphaCheckResponse struct {
	// Callsign of the friendly aircraft requesting the ALPHA CHECK.
	Callsign string
	// Status is true if the ALPHA CHECK was correlated to an aircraft on frequency, otherwise false.
	Status bool
	// Location of the friendly aircraft. If Status is false, this may be nil.
	Location Bullseye
}
