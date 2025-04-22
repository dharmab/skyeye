package controller

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog"
)

type correlation struct {
	Callsign string
	Bearing  bearings.Bearing
	Group    brevity.Group
}

func (c *Controller) correlate(logger zerolog.Logger, callsign string, bearing bearings.Bearing) correlation {
	logger.Debug().Msg("handling request")

	if !bearing.IsMagnetic() {
		logger.Error().Stringer("bearing", bearing).Msg("bearing should be magnetic")
	}

	foundCallsign, trackfile, ok := c.findCallsign(callsign)
	if !ok {
		return correlation{}
	}

	origin := trackfile.LastKnown().Point
	arc := unit.Angle(30) * unit.Degree
	distance := unit.Length(120) * unit.NauticalMile
	nearestGroup := c.scope.FindNearestGroupInSector(
		origin,
		lowestAltitude,
		highestAltitude,
		distance,
		bearing,
		arc,
		c.coalition.Opposite(),
		brevity.FixedWing,
	)

	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found within cone")
		return correlation{
			Callsign: foundCallsign,
			Bearing:  bearing,
		}
	}
	nearestGroup.SetDeclaration(brevity.Hostile)

	logger = logger.With().Stringer("group", nearestGroup).Logger()
	logger.Debug().Msg("hostile group found within cone")
	return correlation{
		Callsign: foundCallsign,
		Bearing:  bearing,
		Group:    nearestGroup,
	}
}
