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
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	targetLocation := geo.PointAtBearingAndDistance(
		requestorTrackfile.LastKnown().Point,
		r.BRA.Bearing().Degrees(),
		r.BRA.Range().Meters(),
	)
	group := c.scope.FindNearestGroupWithBullseye(targetLocation, c.hostileCoalition(), brevity.Airplanes)

	status := group != nil
	logger.Debug().Bool("status", status).Msg("responding to SNAPLOCK request")
	c.out <- brevity.SnaplockResponse{
		Callsign: r.Callsign,
		Status:   status,
		Group:    group,
	}
}
