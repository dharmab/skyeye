package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleRadioCheck implements Controller.HandleRadioCheck.
func (c *controller) HandleRadioCheck(request *brevity.RadioCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")
	var response brevity.RadioCheckResponse
	if foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition); trackfile == nil {
		logger.Debug().Msg("no trackfile found for requestor")
		response.Callsign = request.Callsign
		return
	} else {
		logger.Debug().Msg("found requestor's trackfile")
		response.Callsign = foundCallsign
		response.RadarContact = true
	}
	c.out <- response
}
