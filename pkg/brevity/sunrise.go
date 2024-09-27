package brevity

import (
	"fmt"

	"github.com/martinlindhe/unit"
)

// SunriseCall reports that the GCI is online and ready for requests.
type SunriseCall struct {
	// Frequency which the GCI is listening on.
	Frequencies []unit.Frequency
}

func (c SunriseCall) String() string {
	s := "SUNRISE: frequencies "
	for i, f := range c.Frequencies {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%.3fMHz", f.Megahertz())
	}
	return s
}

// MidnightCall reports that the GCI is offline.
type MidnightCall struct{}
