package brevity

import "github.com/martinlindhe/unit"

// SunriseCall reports that the GCI is online and ready for requests.
type SunriseCall struct {
	// Frequency which the GCI is listening on.
	Frequencies []unit.Frequency
}

// MidnightCall reports that the GCI is offline.
type MidnightCall struct{}
