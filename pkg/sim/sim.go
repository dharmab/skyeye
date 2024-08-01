package sim

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/paulmach/orb"
)

// Sim is the interface for receiving telemetry data from the flight simulator.
type Sim interface {
	// Stream aircraft updates from the sim to the provided channels.
	// The first channel receives updates for active aircraft.
	// The second channel receives messages when an aircraft disappears.
	// This function blocks until the context is cancelled.
	Stream(context.Context, chan<- Updated, chan<- Faded)
	// Bullseye returns the coalition's bullseye center.
	Bullseye() orb.Point
	// Time returns the starting time of the mission.
	// This is useful for looking up magnetic variation.
	Time() time.Time
}

// Updated is a message sent when an aircraft is updated.
type Updated struct {
	// Aircraft contains the aircraft's identity.
	Aircraft trackfiles.Aircraft
	// Frame contains the aircraft's observed position data.
	Frame trackfiles.Frame
}

// Faded is a message sent when an aircraft disappears.
type Faded struct {
	// Timestamp when the aircraft disappeared.
	Timestamp time.Time
	// UnitID of the aircraft that disappeared.
	UnitID uint32
}
