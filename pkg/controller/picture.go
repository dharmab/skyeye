package controller

import (
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandlePicture implements Controller.HandlePicture.
func (c *controller) HandlePicture(request *brevity.PictureRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	logger.Debug().Msg("building picture")
	count, groups := c.scope.GetPicture(conf.DefaultPictureRadius, c.hostileCoalition(), brevity.FixedWing)
	for _, group := range groups {
		group.SetDeclaration(brevity.Hostile)
	}

	logger.Debug().Int("groups", len(groups)).Int("count", count).Msg("sending response")
	c.out <- brevity.PictureResponse{Count: count, Groups: groups}
}
