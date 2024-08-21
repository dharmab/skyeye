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

	c.broadcastPicture(&logger, true)
}

func (c *controller) broadcastPicture(logger *zerolog.Logger, forceBroadcast bool) {
	count, groups := c.scope.GetPicture(conf.DefaultPictureRadius, c.coalition.Opposite(), brevity.FixedWing)
	isPictureClean := count == 0
	for _, group := range groups {
		group.SetDeclaration(brevity.Hostile)
	}

	if c.wasLastPictureClean && isPictureClean && !forceBroadcast {
		logger.Info().Msg("skipping PICTURE broadcast because situation has not changed since last broadcast")
	} else {
		logger.Info().Int("groups", len(groups)).Int("count", count).Msg("broadcasting PICTURE")
		c.out <- brevity.PictureResponse{Count: count, Groups: groups}
	}

	c.pictureBroadcastDeadline = time.Now().Add(c.pictureBroadcastInterval)
	c.wasLastPictureClean = isPictureClean
	logger.Info().Time("deadline", c.pictureBroadcastDeadline).Msg("extended next PICTURE broadcast time")
}
