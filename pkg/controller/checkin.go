package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleCheckIn handles an ambiguous CHECK IN by asking the player to clarify their call.
func (c *Controller) HandleCheckIn(ctx context.Context, request *brevity.CheckInRequest) {
	checkInCounter.Add(ctx, 1)
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	foundCallsign, _, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	c.calls <- NewCall(ctx, brevity.CheckInResponse{Callsign: foundCallsign})
}
