package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleRadioCheck handles a RADIO CHECK by responding to the requesting aircraft.
func (c *Controller) HandleRadioCheck(ctx context.Context, request *brevity.RadioCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")
	var response brevity.RadioCheckResponse
	foundCallsign, _, ok := c.findCallsign(request.Callsign)
	if !ok {
		response.Callsign = request.Callsign
		response.RadarContact = false
	} else {
		response.Callsign = foundCallsign
		response.RadarContact = true
	}
	c.calls <- NewCall(ctx, response)
}
