package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
func (c *Controller) HandleBogeyDope(ctx context.Context, request *brevity.BogeyDopeRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Any("filter", request.Filter).Logger()
	logger.Debug().Msg("handling request")

	foundCallsign, trackfile, ok := c.findCallsign(request.Callsign)
	if !ok {
		c.calls <- NewCall(ctx, brevity.NegativeRadarContactResponse{Callsign: request.Callsign})
		return
	}
	logger = logger.With().Str("callsign", foundCallsign).Logger()

	origin := trackfile.LastKnown().Point
	logger.Debug().Any("origin", origin).Msgf("determined origin point for BOGEY DOPE, lat %s, lon %s", origin.Lat(), origin.Lon())
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
	logger.Debug().Any("braa", nearestGroup.BRAA().Bearing().Degrees()).Msg("determined BRAA for nearest hostile group")
	if nearestGroup.BRAA().Bearing().IsMagnetic() == false {
		logger.Debug().Msg("bearing is true")
	} else if nearestGroup.BRAA().Bearing().IsMagnetic() == true {
		logger.Debug().Msg("bearing is magnetic")
	}
	//logger.Debug().Any("bullseye", nearestGroup.Bullseye().Bearing().Degrees()).Msg("determined Bullseye for nearest hostile group")

	logger.Info().
		Strs("platforms", nearestGroup.Platforms()).
		Str("aspect", string(nearestGroup.Aspect())).
		Msg("found nearest hostile group")
	c.calls <- NewCall(ctx, brevity.BogeyDopeResponse{Callsign: foundCallsign, Group: nearestGroup})
}
