package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
)

// HandleSnaplock implements Controller.HandleSnaplock.
func (c *controller) HandleSnaplock(r *brevity.SnaplockRequest) {
	requestorTrackfile := c.scope.FindCallsign(r.Callsign)
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

	c.out <- brevity.SnaplockResponse{
		Callsign: r.Callsign,
		Status:   group != nil,
		Group:    group,
	}
}
