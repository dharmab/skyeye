package controller

import (
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HandlePicture implements Controller.HandlePicture.
func (c *controller) HandlePicture(request *brevity.PictureRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	c.broadcastPicture(&logger)
}

func (c *controller) broadcastPicture(logger *zerolog.Logger) {
	logger.Debug().Msg("building picture")
	count, groups := c.scope.GetPicture(conf.DefaultPictureRadius, c.hostileCoalition(), brevity.FixedWing)
	for _, group := range groups {
		group.SetDeclaration(brevity.Hostile)
	}

	logger.Info().Int("groups", len(groups)).Int("count", count).Msg("broadcasting PICTURE")
	c.out <- brevity.PictureResponse{Count: count, Groups: groups}

	c.pictureBroadcastDeadline = time.Now().Add(c.pictureBroadcastInterval)
	logger.Info().Time("deadline", c.pictureBroadcastDeadline).Msg("extending next PICTURE broadcast time")
}
