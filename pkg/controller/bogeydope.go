package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleBogeyDope implements Controller.HandleBogeyDope.
func (c *controller) HandleBogeyDope(request *brevity.BogeyDopeRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Any("filter", request.Filter).Logger()
	logger.Debug().Msg("handling request")

	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}

	logger = logger.With().Str("callsign", foundCallsign).Logger()
	logger.Info().Stringer("trackfile", trackfile).Msg("found requestor's trackfile")

	origin := trackfile.LastKnown().Point
	radius := 300 * unit.NauticalMile
	nearestGroup := c.scope.FindNearestGroupWithBRAA(
		origin,
		lowestAltitude,
		highestAltitude,
		radius,
		c.coalition.Opposite(),
		request.Filter,
	)

	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found")
		c.out <- brevity.BogeyDopeResponse{Callsign: foundCallsign, Group: nil}
		return
	}

	nearestGroup.SetDeclaration(brevity.Hostile)
	c.fillInMergeDetails(nearestGroup)

	logger.Info().
		Strs("platforms", nearestGroup.Platforms()).
		Str("aspect", string(nearestGroup.Aspect())).
		Msg("found nearest hostile group")
	c.out <- brevity.BogeyDopeResponse{Callsign: foundCallsign, Group: nearestGroup}
}
