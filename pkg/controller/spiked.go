package controller

import (
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
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
	hostileGroup := c.scope.FindNearestGroupInCone(requestorLocation, r.Bearing, 30, c.hostileCoalition(), brevity.FixedWing)
	friendlyGroup := c.scope.FindNearestGroupInCone(requestorLocation, r.Bearing, 30, c.coalition, brevity.FixedWing)
	resp := brevity.SpikedResponse{
		Callsign: r.Callsign,
		Bearing:  r.Bearing,
	}

	if hostileGroup == nil && friendlyGroup == nil {
		logger.Debug().Msg("no groups found within spike cone")
		resp.Status = false
		resp.Aspect = brevity.UnknownAspect
		resp.Declaration = brevity.Clean
	} else {
		resp.Status = true
		var nearestGroup brevity.Group
		if hostileGroup != nil && friendlyGroup != nil {
			logger.Debug().Msg("both hostile and friendly groups found within spike cone")
			if hostileGroup.BRAA().Range() < friendlyGroup.BRAA().Range() {
				logger.Debug().Msg("hostile group is closer")
				nearestGroup = hostileGroup
				resp.Declaration = brevity.Hostile
			} else {
				logger.Debug().Msg("friendly group is closer")
				nearestGroup = friendlyGroup
				resp.Declaration = brevity.Friendly
			}
			// check if hostile and friendly within 3nm
			hostilePoint := geo.PointAtBearingAndDistance(requestorLocation, hostileGroup.BRAA().Bearing().Degrees(), hostileGroup.BRAA().Range().Meters())
			friendlyPoint := geo.PointAtBearingAndDistance(requestorLocation, friendlyGroup.BRAA().Bearing().Degrees(), friendlyGroup.BRAA().Range().Meters())
			if geo.Distance(hostilePoint, friendlyPoint) < (conf.DefaultMarginRadius).Meters() {
				logger.Debug().Msg("hostile and friendly groups are within 3nm (furball)")
				resp.Declaration = brevity.Furball
				resp.Aspect = brevity.UnknownAspect
			}
		} else if hostileGroup != nil {
			logger.Debug().Msg("hostile group found within spike cone")
			nearestGroup = hostileGroup
			resp.Declaration = brevity.Hostile
			resp.Status = true
		} else {
			logger.Debug().Msg("friendly group found within spike cone")
			nearestGroup = friendlyGroup
			resp.Declaration = brevity.Friendly
		}

		_range := nearestGroup.BRAA().Range()
		resp.Range = &_range
		altitude := nearestGroup.BRAA().Altitude()
		resp.Altitude = &altitude
		resp.Aspect = nearestGroup.BRAA().Aspect()
		resp.Track = nearestGroup.Track()
		resp.Contacts = nearestGroup.Contacts()
	}

	c.out <- resp
}
