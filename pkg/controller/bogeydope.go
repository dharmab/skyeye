package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
func (c *Controller) HandleBogeyDope(ctx context.Context, request *brevity.BogeyDopeRequest) {
	bogeyDopeCounter.Add(ctx, 1)
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Any("filter", request.Filter).Logger()
	logger.Debug().Msg("handling request")

	foundCallsign, trackfile, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	logger = logger.With().Str("callsign", foundCallsign).Logger()

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
		c.calls <- NewCall(ctx, brevity.BogeyDopeResponse{Callsign: foundCallsign, Group: nil})
		return
	}

	nearestGroup.SetDeclaration(brevity.Hostile)
	c.fillInMergeDetails(nearestGroup)

	logger.Info().
		Strs("platforms", nearestGroup.Platforms()).
		Str("aspect", string(nearestGroup.Aspect())).
		Msg("found nearest hostile group")
	c.calls <- NewCall(ctx, brevity.BogeyDopeResponse{Callsign: foundCallsign, Group: nearestGroup})
}
