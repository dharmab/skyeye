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
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Client interface {
	Run(context.Context, *sync.WaitGroup) error
	Bullseye(coalitions.Coalition) (orb.Point, error)
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

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-sCtx.Done():
				log.Info().Msg("stopping time and bullseye updates due to context cancellation")
				return
			case <-ticker.C:
				c.updateTime(source)
				err := c.updateBullseyes(source)
				if err != nil {
					log.Warn().Err(err).Msg("error updating bullseyes")
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping tacview client due to context cancellation")
			return nil
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

func (c *tacviewClient) updateTime(source acmi.ACMI) {
	c.missionTime = source.Time()
}

func (c *tacviewClient) updateBullseyes(source acmi.ACMI) error {
	c.bullseyesLock.Lock()
	defer c.bullseyesLock.Unlock()
	for _, coalition := range []coalitions.Coalition{coalitions.Red, coalitions.Blue} {
		point, err := source.Bullseye(coalition)
		if err != nil {
			return fmt.Errorf("error reading bullseye from ACMI source: %w", err)
		}
		c.bullseyes[coalition] = point
	}
	return nil
}

func (c *tacviewClient) Bullseye(coalition coalitions.Coalition) (orb.Point, error) {
	c.bullseyesLock.RLock()
	defer c.bullseyesLock.RUnlock()
	point, ok := c.bullseyes[coalition]
	if !ok {
		return orb.Point{}, fmt.Errorf("bullseye not found for coalition %d", int(coalition))
	}
	if spatial.IsZero(point) {
		log.Warn().Int("coalition", int(coalition)).Msg("bullseye is set to zero value")
	}
	return point, nil
}
