package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandlePicture implements Controller.HandlePicture.
func (c *controller) HandlePicture(r *brevity.PictureRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")

	tf := c.scope.FindCallsign(r.Callsign)
	if tf == nil {
		logger.Warn().Msg("callsign not found")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}

	logger.Debug().Msg("building picture")
	groups := c.scope.GetPicture(tf.LastKnown().Point, r.Radius, c.hostileCoalition(), brevity.Aircraft)
	for _, g := range groups {
		g.SetDeclaration(brevity.Bandit)
	}

	logger.Debug().Int("groups", len(groups)).Bool("clean", len(groups) == 0).Msg("sending response")
	c.out <- brevity.PictureResponse{Groups: groups}
}
