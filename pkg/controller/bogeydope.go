package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleBogeyDope implements Controller.HandleBogeyDope.
func (c *controller) HandleBogeyDope(request *brevity.BogeyDopeRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Any("filter", request.Filter).Logger()
	logger.Debug().Msg("handling request")
	requestorTrackfile := c.findCallsign(request.Callsign)
	if requestorTrackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}
	logger = logger.With().Str("requestorTrackfile", requestorTrackfile.String()).Logger()
	logger.Info().Msg("found requestor's trackfile")

	requestorLocation := requestorTrackfile.LastKnown().Point
	nearestGroup := c.scope.FindNearestGroupWithBRAA(requestorLocation, c.hostileCoalition(), request.Filter)
	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found")
		c.out <- brevity.BogeyDopeResponse{Callsign: request.Callsign, Group: nil}
		return
	}
	nearestGroup.SetDeclaration(brevity.Hostile)
	logger.Info().
		Any("platforms", nearestGroup.Platforms()).
		Str("aspect", string(nearestGroup.Aspect())).
		Msg("found nearest hostile group")
	c.out <- brevity.BogeyDopeResponse{Callsign: request.Callsign, Group: nearestGroup}
}
