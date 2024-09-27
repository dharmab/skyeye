package brevity

// NegativeRadarContactResponse provides a response when the GCI controller cannot find the caller on the radar scope.
type NegativeRadarContactResponse struct {
	// Callsign of the friendly aircraft that made the request.
	Callsign string
}

func (r NegativeRadarContactResponse) String() string {
	return "NEGATIVE RADAR CONTACT: callsign " + r.Callsign
}
