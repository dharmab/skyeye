package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// HandleDeclare implements Controller.HandleDeclare.
func (c *controller) HandleDeclare(request *brevity.DeclareRequest) {
	logger := log.With().Str("callsign", request.Callsign).Type("type", request).Logger()
	logger.Debug().Msg("handling request")

	logger.Info().
		Float64("bearingDegrees", request.Location.Bearing().Degrees()).
		Float64("distanceNM", request.Location.Distance().NauticalMiles()).
		Float64("altitudeFeet", request.Altitude.Feet()).
		Msg("handling DECLARE request")

	if !request.Location.Bearing().IsMagnetic() {
		logger.Error().Any("bearing", request.Location.Bearing()).Msg("bearing provided to HandleDeclare should be magnetic")
	}

	foundCallsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for requestor")
		c.out <- brevity.NegativeRadarContactResponse{Callsign: request.Callsign}
		return
	}

	aoi := geo.PointAtBearingAndDistance(
		c.scope.GetBullseye(),
		request.Location.Bearing().True(c.scope.Declination(c.scope.GetBullseye())).Degrees(),
		request.Location.Distance().Meters(),
	)
	radius := 7 * unit.NauticalMile // TODO reduce to 3 when magvar is available
	altitudeMargin := unit.Length(5000) * unit.Foot
	minAltitude := request.Altitude - altitudeMargin
	maxAltitude := request.Altitude + altitudeMargin
	friendlyGroups := c.scope.FindNearbyGroupsWithBullseye(aoi, minAltitude, maxAltitude, radius, c.coalition, brevity.Aircraft)
	hostileGroups := c.scope.FindNearbyGroupsWithBullseye(aoi, minAltitude, maxAltitude, radius, c.hostileCoalition(), brevity.Aircraft)
	logger.Debug().Int("friendly", len(friendlyGroups)).Int("hostile", len(hostileGroups)).Msg("queried groups near declared location")

	response := brevity.DeclareResponse{Callsign: foundCallsign}
	if len(friendlyGroups)+len(hostileGroups) == 0 {
		logger.Debug().Msg("no groups found")
		response.Declaration = brevity.Clean
	} else if len(friendlyGroups) > 0 && len(hostileGroups) == 0 {
		logger.Debug().Msg("friendly groups found")
		response.Declaration = brevity.Friendly
		response.Group = friendlyGroups[0]
	} else if len(friendlyGroups) == 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("hostile groups found")
		response.Declaration = brevity.Hostile
		response.Group = hostileGroups[0]
	} else if len(friendlyGroups) > 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("both friendly and hostile groups found")
		response.Declaration = brevity.Furball
	}

	if response.Group != nil {
		response.Group.SetDeclaration(response.Declaration)
	}

	logger.Debug().Any("declaration", response.Declaration).Msg("responding to DECLARE request")
	c.out <- response
}
