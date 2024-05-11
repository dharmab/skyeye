package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// HandleBogeyDope implements Controller.HandleBogeyDope.
func (c *controller) HandleBogeyDope(r *brevity.BogeyDopeRequest) {
	requestorTrackfile := c.scope.FindCallsign(r.Callsign)
	if requestorTrackfile == nil {
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	requestorLocation := requestorTrackfile.LastKnown().Point
	nearestGroup := c.scope.FindNearestGroupWithBRAA(requestorLocation, c.hostileCoalition(), r.Filter)
	c.out <- brevity.BogeyDopeResponse{Callsign: r.Callsign, Group: nearestGroup}
}
