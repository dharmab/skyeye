package sim

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/paulmach/orb"
)

type Sim interface {
	// Stream aircraft updates from the sim to the provided channels.
	// The first channel receives updates for active aircraft.
	// The second channel receives messages when an aircraft disappears.
	// This function blocks until the context is cancelled.
	Stream(context.Context, chan<- Updated, chan<- Faded)
	// Bullseye returns the coalition's bullseye center.
	Bullseye() orb.Point
}

type Updated struct {
	Aircraft trackfile.Aircraft
	Frame    trackfile.Frame
}

type Faded struct {
	Timestamp time.Time
	UnitID    uint32
}
