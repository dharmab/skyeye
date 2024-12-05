package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleShopping handles a SHOPPING request... by not implementing it, since it's not an air-to-air call!
func (c *Controller) HandleShopping(ctx context.Context, request *brevity.ShoppingRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	foundCallsign, _, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	c.calls <- NewCall(ctx, brevity.ShoppingResponse{Callsign: foundCallsign})
}
