package client

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Client interface {
	Run(context.Context) error
	Close() error
}

type tacviewClient struct {
	coalition      coalitions.Coalition
	updates        chan<- sim.Updated
	fades          chan<- sim.Faded
	bullseyes      chan<- orb.Point
	updateInterval time.Duration
}

func (c *tacviewClient) run(ctx context.Context, source acmi.ACMI) error {
	go func() {
		err := source.Start(ctx)
		if err != nil {
			log.Error().Err(err).Msg("error starting ACMI client")
		}
	}()
	go source.Stream(ctx, c.updates, c.fades)
	go func() {
		ticker := time.NewTicker(c.updateInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.bullseyes <- source.Bullseye()
			}
		}
	}()

	<-ctx.Done()
	return nil
}
