package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *controller) HandleCheckIn(ctx context.Context, request *brevity.CheckInRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	foundCallsign, _, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	c.calls <- NewCall(ctx, brevity.CheckInResponse{Callsign: foundCallsign})
}
