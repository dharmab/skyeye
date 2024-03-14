package brevity

import "github.com/martinlindhe/unit"

// SunriseCall reports that the GCI is online and ready for requests.
type SunriseCall interface {
	Sunrise()
	// Frequency which the GCI is listening on.
	Frequency() unit.Frequency
}
