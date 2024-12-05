package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleTripwire handles a TRIPWIRE... by not implementing it LOL!
func (c *Controller) HandleTripwire(ctx context.Context, request *brevity.TripwireRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	foundCallsign, _, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	c.calls <- NewCall(ctx, brevity.TripwireResponse{Callsign: foundCallsign})
}
