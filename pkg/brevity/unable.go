package brevity

// UnableToUnderstandRequest provides a response when the GCI controller cannot understand the caller's request because either the caller's callsign or the request itself is unclear.
type UnableToUnderstandRequest struct {
	// Callsign of the friendly aircraft that made the request.
	// If the callsign was unclear, this field will be empty.
	Callsign string
}

// SayAgainResponse is a generic response asking the caller to repeat their last transmission.
type SayAgainResponse struct {
	// Callsign of the friendly aircraft that made the request.
	// This may be empty if the GCI is unsure of the caller's identity.
	Callsign string
}
