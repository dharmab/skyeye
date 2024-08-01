package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleSpiked implements Controller.HandleSpiked.
func (c *controller) HandleSpiked(request *brevity.SpikedRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Float64("bearing", request.Bearing.Degrees()).Logger()
	logger.Debug().Msg("handling request")

	trackfile := c.scope.FindCallsign(request.Callsign)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}
	origin := trackfile.LastKnown().Point
	arc := unit.Angle(30) * unit.Degree
	nearestGroup := c.scope.FindNearestGroupInCone(origin, request.Bearing, arc, c.hostileCoalition(), brevity.FixedWing)

	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found within spike cone")
		c.out <- brevity.SpikedResponse{Callsign: request.Callsign, Status: false}
		return
	}

	logger = logger.With().Any("group", nearestGroup).Logger()
	logger.Debug().Msg("hostile group found within spike cone")
	c.out <- brevity.SpikedResponse{
		Callsign:    request.Callsign,
		Status:      true,
		Bearing:     request.Bearing,
		Range:       nearestGroup.BRAA().Range(),
		Altitude:    nearestGroup.BRAA().Altitude(),
		Aspect:      nearestGroup.BRAA().Aspect(),
		Track:       nearestGroup.Track(),
		Declaration: brevity.Hostile,
		Contacts:    nearestGroup.Contacts(),
	}
}
