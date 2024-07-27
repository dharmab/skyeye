package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleAlphaCheck implements [Controller.HandleAlphaCheck].
func (c *controller) HandleAlphaCheck(r *brevity.AlphaCheckRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	tf := c.findCallsign(r.Callsign)
	if tf == nil {
		logger.Debug().Msg("no trackfile found for requestor")
		c.out <- brevity.AlphaCheckResponse{
			Callsign: r.Callsign,
			Status:   false,
		}
		return
	}
	logger.Debug().Msg("found requestor's trackfile")
	location := tf.Bullseye(c.scope.GetBullseye())
	c.out <- brevity.AlphaCheckResponse{
		Callsign: r.Callsign,
		Status:   true,
		Location: location,
	}
}
