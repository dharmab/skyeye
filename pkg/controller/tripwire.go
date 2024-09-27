package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *controller) HandleTripwire(ctx context.Context, request *brevity.TripwireRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		log.Debug().Msg("no trackfile found for requestor")
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	c.calls <- NewCall(ctx, brevity.TripwireResponse{Callsign: foundCallsign})
}
