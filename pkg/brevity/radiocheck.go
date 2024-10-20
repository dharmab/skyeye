package brevity

// RadioCheckRequest is a request for a RADIO CHECK.
type RadioCheckRequest struct {
	// Callsign of the friendly aircraft requesting the RADIO CHECK.
	Callsign string
}

func (r RadioCheckRequest) String() string {
	return "RADIO CHECK for " + r.Callsign
}

// RadioCheckResponse is a response to a RADIO CHECK.
type RadioCheckResponse struct {
	// Callsign of the friendly aircraft requesting the RADIO CHECK.
	// If the callsign was misheard, this may not be the actual callsign of any actual aircraft.
	Callsign string
	// RadarContact indicates whether the callsign was found on the radar scope.
	RadarContact bool
}
