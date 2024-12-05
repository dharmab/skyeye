package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/paulmach/orb"
)

// Client reads telemetry data from a data source.
type Client interface {
	// Run reads telemetry data.
	Run(context.Context) error
	// Stream publishes telemetry data to the given channels.
	Stream(context.Context, *sync.WaitGroup, chan<- sim.Started, chan<- sim.Updated, chan<- sim.Faded)
	// Bullseye returns the position of the given coalition's bullseye.
	Bullseye(coalitions.Coalition) (orb.Point, error)
	// Time returns the current mission time.
	Time() time.Time
}
