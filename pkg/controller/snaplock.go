package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// HandleSnaplock implements Controller.HandleSnaplock.
func (c *controller) HandleSnaplock(request *brevity.SnaplockRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	if !request.BRA.Bearing().IsMagnetic() {
		logger.Error().Any("bearing", request.BRA.Bearing()).Msg("bearing provided to HandleSnaplock should be magnetic")
	}

	requestorTrackfile := c.findCallsign(request.Callsign)
	if requestorTrackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}

	targetLocation := geo.PointAtBearingAndDistance(
		requestorTrackfile.LastKnown().Point,
		request.BRA.Bearing().Degrees(),
		request.BRA.Range().Meters(),
	)
	group := c.scope.FindNearestGroupWithBullseye(targetLocation, c.hostileCoalition(), brevity.Aircraft)

	status := group != nil
	if !status {
		logger.Debug().Msg("no hostile groups found")
	} else {
		logger.Debug().Msg("found nearest hostile group")
	}
	c.out <- brevity.SnaplockResponse{
		Callsign: request.Callsign,
		Status:   status,
		Group:    group,
	}
}
