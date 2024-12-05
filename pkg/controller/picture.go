package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// HandlePicture handles a PICTURE by reporting a tactical air picture.
func (c *Controller) HandlePicture(ctx context.Context, request *brevity.PictureRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	c.broadcastPicture(ctx, &logger, true)
}

func (c *Controller) broadcastPicture(ctx context.Context, logger *zerolog.Logger, forceBroadcast bool) {
	if !forceBroadcast {
		if c.srsClient.ClientsOnFrequency() == 0 {
			logger.Debug().Msg("skipping PICTURE broadcast because no clients are on frequency")
			return
		}
		c.scope.WaitUntilFadesResolve(ctx)
	}
	count, groups := c.scope.GetPicture(conf.DefaultPictureRadius, c.coalition.Opposite(), brevity.FixedWing)
	isPictureClean := count == 0
	for _, group := range groups {
		group.SetDeclaration(brevity.Hostile)
		c.fillInMergeDetails(group)
	}

	if c.wasLastPictureClean && isPictureClean && !forceBroadcast {
		logger.Info().Msg("skipping PICTURE broadcast because situation has not changed since last broadcast")
	} else {
		logger.Info().Int("groups", len(groups)).Int("count", count).Msg("broadcasting PICTURE")
		c.calls <- NewCall(ctx, brevity.PictureResponse{Count: count, Groups: groups})
	}

	c.pictureBroadcastDeadline = time.Now().Add(c.pictureBroadcastInterval)
	c.wasLastPictureClean = isPictureClean
	logger.Info().Time("deadline", c.pictureBroadcastDeadline).Msg("extended next PICTURE broadcast time")
}
