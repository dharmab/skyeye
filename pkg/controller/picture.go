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
	count, groups := c.scope.GetPicture(tf.LastKnown().Point, r.Radius, c.hostileCoalition(), brevity.FixedWing)
	for _, g := range groups {
		g.SetDeclaration(brevity.Bandit)
	}

	logger.Debug().Int("groups", len(groups)).Int("count", count).Msg("sending response")
	c.out <- brevity.PictureResponse{Count: count, Groups: groups}
}
