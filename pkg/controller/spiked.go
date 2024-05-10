package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
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
	resp := &brevity.SpikedResponse{
		Callsign: r.Callsign,
		Bearing:  r.Bearing,
	}

	if hostileGroup == nil && friendlyGroup == nil {
		resp.Status = false
		resp.Aspect = brevity.UnknownAspect
		resp.Declaration = brevity.Clean
	} else {
		resp.Status = true
		var nearestGroup brevity.Group
		if hostileGroup != nil && friendlyGroup != nil {
			if hostileGroup.BRAA().Range() < friendlyGroup.BRAA().Range() {
				nearestGroup = hostileGroup
			} else {
				nearestGroup = friendlyGroup
			}
			// check if hostile and friendly within 3nm
			hostilePoint := geo.PointAtBearingAndDistance(requestorLocation, hostileGroup.BRAA().Bearing().Degrees(), hostileGroup.BRAA().Range().Meters())
			friendlyPoint := geo.PointAtBearingAndDistance(requestorLocation, friendlyGroup.BRAA().Bearing().Degrees(), friendlyGroup.BRAA().Range().Meters())
			if planar.Distance(hostilePoint, friendlyPoint) < (3 * unit.NauticalMile).Meters() {
				resp.Declaration = brevity.Furball
			}
		} else if hostileGroup != nil {
			nearestGroup = hostileGroup
			resp.Declaration = brevity.Hostile
		} else {
			nearestGroup = friendlyGroup
			resp.Declaration = brevity.Friendly
		}

		_range := nearestGroup.BRAA().Range()
		resp.Range = &_range
		altitude := nearestGroup.BRAA().Altitude()
		resp.Altitude = &altitude
		resp.Aspect = nearestGroup.Aspect()
	}

	c.out <- resp
}
