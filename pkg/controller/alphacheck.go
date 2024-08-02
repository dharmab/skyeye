package controller

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleAlphaCheck implements [Controller.HandleAlphaCheck].
func (c *controller) HandleAlphaCheck(request *brevity.AlphaCheckRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")
	trackfile := c.findCallsign(request.Callsign)
	if trackfile == nil {
		logger.Debug().Msg("no trackfile found for requestor")
		c.out <- brevity.AlphaCheckResponse{
			Callsign: request.Callsign,
			Status:   false,
		}
		return
	}
	logger.Debug().Msg("found requestor's trackfile")
	declination, err := bearings.Declination(c.bullseye, c.missionTime)
	if err != nil {
		logger.Error().Err(err).Msg("failed to calculate declination")
	}
	location := trackfile.Bullseye(c.bullseye, declination)
	c.out <- brevity.AlphaCheckResponse{
		Callsign: request.Callsign,
		Status:   true,
		Location: location,
	}
}
