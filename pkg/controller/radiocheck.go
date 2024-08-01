package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleRadioCheck implements Controller.HandleRadioCheck.
func (c *controller) HandleRadioCheck(r *brevity.RadioCheckRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	trackfile := c.findCallsign(r.Callsign)
	status := trackfile != nil
	if !status {
		logger.Debug().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	logger.Debug().Msg("found requestor's trackfile")
	c.out <- brevity.RadioCheckResponse{Callsign: r.Callsign}
}
