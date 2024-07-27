package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// HandleSnaplock implements Controller.HandleSnaplock.
func (c *controller) HandleSnaplock(r *brevity.SnaplockRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")

	requestorTrackfile := c.findCallsign(r.Callsign)
	if requestorTrackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}

	targetLocation := geo.PointAtBearingAndDistance(
		requestorTrackfile.LastKnown().Point,
		r.BRA.Bearing().Degrees(),
		r.BRA.Range().Meters(),
	)
	group := c.scope.FindNearestGroupWithBullseye(targetLocation, c.hostileCoalition(), brevity.Aircraft)

	status := group != nil
	if !status {
		logger.Debug().Msg("no hostile groups found")
	} else {
		logger.Debug().Msg("found nearest hostile group")
	}
	c.out <- brevity.SnaplockResponse{
		Callsign: r.Callsign,
		Status:   status,
		Group:    group,
	}
}
