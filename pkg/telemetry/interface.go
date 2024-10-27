package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/paulmach/orb"
)

type Client interface {
	Run(context.Context, *sync.WaitGroup) error
	Stream(context.Context, *sync.WaitGroup, chan<- sim.Started, chan<- sim.Updated, chan<- sim.Faded)
	Bullseye(coalitions.Coalition) (orb.Point, error)
	Time() time.Time
}
