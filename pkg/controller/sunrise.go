package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/lithammer/shortuuid/v3"
	"github.com/martinlindhe/unit"
)

func (c *Controller) broadcastSunrise(ctx context.Context) {
	frequencies := make([]unit.Frequency, 0)
	for _, rf := range c.srsClient.Frequencies() {
		frequencies = append(frequencies, rf.Frequency)
	}
	c.calls <- NewCall(traces.WithTraceID(ctx, shortuuid.New()), brevity.SunriseCall{Frequencies: frequencies})
}
