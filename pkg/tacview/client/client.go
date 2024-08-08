// client contains clients to stream ACMI data from a local or remote source.
package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

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
	}
}

func (c *tacviewClient) stream(ctx context.Context, wg *sync.WaitGroup, source acmi.ACMI) error {
	sCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		source.Stream(sCtx, c.updates, c.fades)
	}()

	ticker := time.NewTicker(c.updateInterval)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping tacview client due to context cancellation")
			return nil
		case <-ticker.C:
			for _, coalition := range []coalitions.Coalition{coalitions.Red, coalitions.Blue} {
				c.bullseyesLock.Lock()
				c.bullseyes[coalition] = source.Bullseye(coalition)
				c.bullseyesLock.Unlock()
				c.missionTime = source.Time()
			}

		default:
			err := source.Run(ctx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Info().Msg("ACMI source closed")
					return fmt.Errorf("ACMI source closed: %w", err)
				} else {
					log.Error().Err(err).Msg("error starting ACMI client")
					return err
				}
			}
		}
	}
}

func (c *tacviewClient) Bullseye(coalition coalitions.Coalition) orb.Point {
	c.bullseyesLock.RLock()
	defer c.bullseyesLock.RUnlock()
	return c.bullseyes[coalition]
}
