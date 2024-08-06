// client contains clients to stream ACMI data from a local or remote source.
package client

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Client interface {
	Run(context.Context, *sync.WaitGroup) error
	Bullseye(coalitions.Coalition) orb.Point
	Time() time.Time
	Close() error
}

type tacviewClient struct {
	updates        chan<- sim.Updated
	fades          chan<- sim.Faded
	updateInterval time.Duration
	bullseyes      map[coalitions.Coalition]orb.Point
	bullseyesLock  sync.RWMutex
	missionTime    time.Time
}

func newTacviewClient(updates chan<- sim.Updated, fades chan<- sim.Faded, updateInterval time.Duration) *tacviewClient {
	return &tacviewClient{
		updates:        updates,
		fades:          fades,
		updateInterval: updateInterval,
		bullseyes:      map[coalitions.Coalition]orb.Point{},
		missionTime:    conf.InitialTime,
	}
}

func (c *tacviewClient) run(ctx context.Context, wg *sync.WaitGroup, source acmi.ACMI) error {
	c.missionTime = conf.InitialTime
	wg.Add(3)
	go func() {
		defer wg.Done()
		err := source.Start(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error starting ACMI client")
		}
	}()
	go func() {
		defer wg.Done()
		source.Stream(ctx, c.updates, c.fades)
	}()
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(c.updateInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, coalition := range []coalitions.Coalition{coalitions.Red, coalitions.Blue} {
					c.bullseyesLock.Lock()
					c.bullseyes[coalition] = source.Bullseye(coalition)
					c.bullseyesLock.Unlock()
					c.missionTime = source.Time()
				}
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func (c *tacviewClient) Bullseye(coalition coalitions.Coalition) orb.Point {
	c.bullseyesLock.RLock()
	defer c.bullseyesLock.RUnlock()
	return c.bullseyes[coalition]
}
