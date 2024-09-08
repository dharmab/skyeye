package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// HandleSnaplock implements Controller.HandleSnaplock.
func (c *controller) HandleSnaplock(request *brevity.SnaplockRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Any("request", request).Logger()

	logger.
		Info().
		Float64("bearing", request.BRA.Bearing().Rounded().Degrees()).
		Float64("range", request.BRA.Range().NauticalMiles()).
		Float64("altitude", request.BRA.Altitude().Feet()).
		Msg("received request")

	if !request.BRA.Bearing().IsMagnetic() {
		logger.Error().Stringer("bearing", request.BRA.Bearing()).Msg("bearing provided to HandleSnaplock should be magnetic")
	}

	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}

	origin := trackfile.LastKnown().Point
	pointOfInterest := spatial.PointAtBearingAndDistance(
		origin,
		request.BRA.Bearing().True(c.scope.Declination(origin)),
		request.BRA.Range(),
	)
	radius := 10 * unit.NauticalMile // TODO reduce to 3 when magvar is available
	altitudeMargin := unit.Length(5000) * unit.Foot
	minAltitude := request.BRA.Altitude() - altitudeMargin
	maxAltitude := request.BRA.Altitude() + altitudeMargin
	friendlyGroups := c.scope.FindNearbyGroupsWithBRAA(
		origin,
		pointOfInterest,
		minAltitude,
		maxAltitude,
		radius,
		c.coalition,
		brevity.Aircraft,
		[]uint64{trackfile.Contact.ID},
	)
	hostileGroups := c.scope.FindNearbyGroupsWithBRAA(
		origin,
		pointOfInterest,
		minAltitude,
		maxAltitude,
		radius,
		c.coalition.Opposite(),
		brevity.Aircraft,
		[]uint64{trackfile.Contact.ID},
	)

	response := brevity.SnaplockResponse{Callsign: foundCallsign}

	// TODO better algorithm for picking from multiple possible groups
	if len(friendlyGroups)+len(hostileGroups) == 0 {
		response.Declaration = brevity.Clean
	} else if len(friendlyGroups) > 0 && len(hostileGroups) == 0 {
		response.Declaration = brevity.Friendly
		response.Group = friendlyGroups[0]
	} else if len(friendlyGroups) == 0 && len(hostileGroups) > 0 {
		response.Declaration = brevity.Hostile
		response.Group = hostileGroups[0]
		for _, group := range hostileGroups {
			if group.Aspect() == brevity.Hot {
				response.Group = group
				break
			}
		}
	} else if len(friendlyGroups) > 0 && len(hostileGroups) > 0 {
		response.Declaration = brevity.Furball
	}

	if response.Group != nil {
		response.Group.SetDeclaration(response.Declaration)
	}

	c.out <- response
}
