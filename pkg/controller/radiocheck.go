package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleRadioCheck implements Controller.HandleRadioCheck.
func (c *controller) HandleRadioCheck(request *brevity.RadioCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")
	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		logger.Debug().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}
	logger.Debug().Msg("found requestor's trackfile")
	c.out <- brevity.RadioCheckResponse{Callsign: foundCallsign}
}
