package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleRadioCheck implements Controller.HandleRadioCheck.
func (c *controller) HandleRadioCheck(r *brevity.RadioCheckRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	tf := c.findCallsign(r.Callsign)
	status := tf != nil
	logger.Debug().Bool("status", status).Msg("responding to RADIO CHECK request")
	c.out <- brevity.RadioCheckResponse{Callsign: r.Callsign, Status: status}
}
