// package sim provides an inteface for receiving telemetry data from DCS World
package sim

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
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
	Bullseye(coalitions.Coalition) (orb.Point, error)
	// Time returns the starting time of the mission.
	// This is useful for looking up magnetic variation.
	Time() time.Time
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
	// Real-time timestamp when the aircraft disappeared.
	Timestamp time.Time
	// Mission time when the aircraft disappeared.
	MissionTimestamp time.Time
	// ID of the aircraft that disappeared.
	ID uint64
}
