package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleAlphaCheck implements [Controller.HandleAlphaCheck].
func (c *controller) HandleAlphaCheck(request *brevity.AlphaCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")
	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		logger.Debug().Msg("no trackfile found for requestor")
		c.out <- brevity.AlphaCheckResponse{
			Callsign: request.Callsign,
			Status:   false,
		}
		return
	}
	logger.Debug().Msg("found requestor's trackfile")
	bullseye := c.scope.Bullseye(trackfile.Contact.Coalition)
	location := trackfile.Bullseye(bullseye)
	c.out <- brevity.AlphaCheckResponse{
		Callsign: foundCallsign,
		Status:   true,
		Location: location,
	}
}
