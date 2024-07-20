package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleAlphaCheck implements Controller.HandleAlphaCheck.
func (c *controller) HandleAlphaCheck(r *brevity.AlphaCheckRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	tf := c.findCallsign(r.Callsign)
	if tf == nil {
		c.out <- brevity.AlphaCheckResponse{
			Callsign: r.Callsign,
			Status:   false,
		}
		return
	}
	location := tf.Bullseye(c.scope.GetBullseye(c.coalition).Point)
	c.out <- brevity.AlphaCheckResponse{
		Callsign: r.Callsign,
		Status:   true,
		Location: location,
	}
}
