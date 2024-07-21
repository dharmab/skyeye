package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleBogeyDope implements Controller.HandleBogeyDope.
func (c *controller) HandleBogeyDope(r *brevity.BogeyDopeRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	requestorTrackfile := c.findCallsign(r.Callsign)
	if requestorTrackfile == nil {
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	requestorLocation := requestorTrackfile.LastKnown().Point
	nearestGroup := c.scope.FindNearestGroupWithBRAA(requestorLocation, c.hostileCoalition(), r.Filter)
	nearestGroup.SetDeclaration(brevity.Hostile)
	c.out <- brevity.BogeyDopeResponse{Callsign: r.Callsign, Group: nearestGroup}
}
