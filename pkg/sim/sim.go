package sim

import (
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

// Started is a message sent when a new mission starts.
type Started struct {
}

// Updated is a message sent when an aircraft is updated.
type Updated struct {
	// Labels contains the aircraft's identity.
	Labels trackfiles.Labels
	// Frame contains the aircraft's observed position data.
	Frame trackfiles.Frame
}

// Faded is a message sent when an aircraft disappears.
type Faded struct {
	// ID of the aircraft that disappeared.
	ID uint64
}
