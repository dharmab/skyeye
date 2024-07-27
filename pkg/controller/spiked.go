package controller

import (
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// HandleSpiked implements Controller.HandleSpiked.
func (c *controller) HandleSpiked(r *brevity.SpikedRequest) {
	requestorTrackfile := c.scope.FindCallsign(r.Callsign)
	if requestorTrackfile == nil {
		c.out <- brevity.NegativeRadarContactResponse{Callsign: r.Callsign}
		return
	}
	requestorLocation := requestorTrackfile.LastKnown().Point
	hostileGroup := c.scope.FindNearestGroupInCone(requestorLocation, r.Bearing, 30, c.hostileCoalition(), brevity.Airplanes)
	friendlyGroup := c.scope.FindNearestGroupInCone(requestorLocation, r.Bearing, 30, c.coalition, brevity.Airplanes)
	resp := brevity.SpikedResponse{
		Callsign: r.Callsign,
		Bearing:  r.Bearing,
	}

	if hostileGroup == nil && friendlyGroup == nil {
		log.Debug().Msg("no groups found within spike cone")
		resp.Status = false
		resp.Aspect = brevity.UnknownAspect
		resp.Declaration = brevity.Clean
	} else {
		resp.Status = true
		var nearestGroup brevity.Group
		if hostileGroup != nil && friendlyGroup != nil {
			log.Debug().Msg("both hostile and friendly groups found within spike cone")
			if hostileGroup.BRAA().Range() < friendlyGroup.BRAA().Range() {
				log.Debug().Msg("hostile group is closer")
				nearestGroup = hostileGroup
				resp.Declaration = brevity.Hostile
			} else {
				log.Debug().Msg("friendly group is closer")
				nearestGroup = friendlyGroup
				resp.Declaration = brevity.Friendly
			}
			// check if hostile and friendly within 3nm
			hostilePoint := geo.PointAtBearingAndDistance(requestorLocation, hostileGroup.BRAA().Bearing().Degrees(), hostileGroup.BRAA().Range().Meters())
			friendlyPoint := geo.PointAtBearingAndDistance(requestorLocation, friendlyGroup.BRAA().Bearing().Degrees(), friendlyGroup.BRAA().Range().Meters())
			if geo.Distance(hostilePoint, friendlyPoint) < (conf.DefaultMarginRadius).Meters() {
				log.Debug().Msg("hostile and friendly groups are within 3nm (furball)")
				resp.Declaration = brevity.Furball
				resp.Aspect = brevity.UnknownAspect
			}
		} else if hostileGroup != nil {
			log.Debug().Msg("hostile group found within spike cone")
			nearestGroup = hostileGroup
			resp.Declaration = brevity.Hostile
			resp.Status = true
		} else {
			log.Debug().Msg("friendly group found within spike cone")
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
