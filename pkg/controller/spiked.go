package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleSpiked implements Controller.HandleSpiked.
func (c *controller) HandleSpiked(r *brevity.SpikedRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Float64("bearing", r.Bearing.Degrees()).Logger()
	logger.Debug().Msg("handling request")

	requestorTrackfile := c.scope.FindCallsign(r.Callsign)
	if requestorTrackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	requestorLocation := requestorTrackfile.LastKnown().Point
	arc := unit.Angle(30) * unit.Degree
	nearestGroup := c.scope.FindNearestGroupInCone(requestorLocation, r.Bearing, arc, c.hostileCoalition(), brevity.FixedWing)

	if nearestGroup == nil {
		logger.Info().Msg("no hostile groups found within spike cone")
		c.out <- brevity.SpikedResponse{Callsign: r.Callsign, Status: false}
		return
	}

	logger = logger.With().Any("group", nearestGroup).Logger()
	logger.Debug().Msg("hostile group found within spike cone")
	c.out <- brevity.SpikedResponse{
		Callsign:    r.Callsign,
		Status:      true,
		Bearing:     r.Bearing,
		Range:       nearestGroup.BRAA().Range(),
		Altitude:    nearestGroup.BRAA().Altitude(),
		Aspect:      nearestGroup.BRAA().Aspect(),
		Track:       nearestGroup.Track(),
		Declaration: brevity.Hostile,
		Contacts:    nearestGroup.Contacts(),
	}
}
