package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandlePicture implements Controller.HandlePicture.
func (c *controller) HandlePicture(request *brevity.PictureRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	trackfile := c.scope.FindCallsign(request.Callsign)
	if trackfile == nil {
		logger.Warn().Msg("callsign not found")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}

	logger.Debug().Msg("building picture")
	count, groups := c.scope.GetPicture(trackfile.LastKnown().Point, request.Radius, c.hostileCoalition(), brevity.FixedWing)
	for _, g := range groups {
		g.SetDeclaration(brevity.Bandit)
	}

	logger.Debug().Int("groups", len(groups)).Int("count", count).Msg("sending response")
	c.out <- brevity.PictureResponse{Count: count, Groups: groups}
}
