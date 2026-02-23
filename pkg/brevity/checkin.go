package brevity

// CheckInRequest is ambiguous.
type CheckInRequest struct {
	Callsign string
}

func (r CheckInRequest) String() string {
	return "CHECK-IN for " + r.Callsign
}

// CheckInResponse is sarcasm.
type CheckInResponse struct {
	Callsign string
}
