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

	if !request.Location.Bearing().IsMagnetic() {
		logger.Error().Any("bearing", request.Location.Bearing()).Msg("bearing provided to HandleDeclare should be magnetic")
	}

	location := geo.PointAtBearingAndDistance(
		c.bullseye,
		request.Location.Bearing().Degrees(),
		request.Location.Distance().Meters(),
	)
	radius := 10 * unit.NauticalMile // TODO reduce to 3 when magvar is available
	friendlyGroups := c.scope.FindNearbyGroups(location, radius, c.coalition, brevity.Aircraft)
	hostileGroups := c.scope.FindNearbyGroups(location, radius, c.hostileCoalition(), brevity.Aircraft)
	logger.Debug().Int("friendly", len(friendlyGroups)).Int("hostile", len(hostileGroups)).Msg("queried groups near delcared location")

	response := brevity.DeclareResponse{Callsign: request.Callsign}
	if len(friendlyGroups)+len(hostileGroups) == 0 {
		logger.Debug().Msg("no groups found")
		response.Declaration = brevity.Clean
	} else if len(friendlyGroups) > 0 && len(hostileGroups) == 0 {
		logger.Debug().Msg("friendly groups found")
		response.Declaration = brevity.Friendly
		response.Group = friendlyGroups[0]
	} else if len(friendlyGroups) == 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("hostile groups found")
		response.Group = hostileGroups[0]
	} else if len(friendlyGroups) > 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("both friendly and hostile groups found")
		response.Declaration = brevity.Furball
	}

	if response.Group != nil {
		response.Group.SetDeclaration(response.Declaration)
	}

	logger.Debug().Interface("declaration", response.Declaration).Msg("responding to DECLARE request")
	c.out <- response
}
