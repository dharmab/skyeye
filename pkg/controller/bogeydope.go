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
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	logger = logger.With().Str("requestorTrackfile", requestorTrackfile.String()).Logger()
	logger.Info().Msg("found requestor's trackfile")
	requestorLocation := requestorTrackfile.LastKnown().Point
	nearestGroup := c.scope.FindNearestGroupWithBRAA(requestorLocation, c.hostileCoalition(), r.Filter)
	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found")
		c.out <- brevity.BogeyDopeResponse{Callsign: r.Callsign, Group: nil}
		return
	}
	nearestGroup.SetDeclaration(brevity.Hostile)
	logger.Info().
		Any("platforms", nearestGroup.Platforms()).
		Msg("found nearest hostile group")
	c.out <- brevity.BogeyDopeResponse{Callsign: r.Callsign, Group: nearestGroup}
}
