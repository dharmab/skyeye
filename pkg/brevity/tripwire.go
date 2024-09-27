package brevity

// TripwireRequest does not exist.
type TripwireRequest struct {
	Callsign string
}

func (r TripwireRequest) String() string {
	return "TRIPWIRE for " + r.Callsign
}

// TripwireResponse is reeducation.
type TripwireResponse struct {
	Callsign string
}

func (r TripwireResponse) String() string {
	return "TRIPWIRE: callsign " + r.Callsign
}
