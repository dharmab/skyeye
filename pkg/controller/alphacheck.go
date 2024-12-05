package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleAlphaCheck handles an ALPHA CHECK by reporting the position of the requesting aircraft.
func (c *Controller) HandleAlphaCheck(ctx context.Context, request *brevity.AlphaCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	foundCallsign, trackfile, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.AlphaCheckResponse{
			Callsign: request.Callsign,
			Status:   false,
		})
		return
	}
	bullseye := c.scope.Bullseye(trackfile.Contact.Coalition)
	location := trackfile.Bullseye(bullseye)
	c.calls <- NewCall(ctx, brevity.AlphaCheckResponse{
		Callsign: foundCallsign,
		Status:   true,
		Location: location,
	})
}
