package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// HandleDeclare implements Controller.HandleDeclare.
func (c *controller) HandleDeclare(r *brevity.DeclareRequest) {
	logger := log.With().Str("callsign", r.Callsign).Type("type", r).Logger()
	logger.Debug().Msg("handling request")
	bullseye := c.scope.GetBullseye()
	location := geo.PointAtBearingAndDistance(
		bullseye,
		r.Location.Bearing().Degrees(),
		r.Location.Distance().Meters(),
	)
	friendlyGroups := c.scope.FindNearbyGroups(location, c.coalition, brevity.Airplanes)
	hostileGroups := c.scope.FindNearbyGroups(location, c.hostileCoalition(), brevity.Airplanes)
	logger.Debug().Int("friendly", len(friendlyGroups)).Int("hostile", len(hostileGroups)).Msg("queried groups near delcared location")

	declaration := brevity.Unable
	var specificGroup brevity.Group
	if len(friendlyGroups)+len(hostileGroups) == 0 {
		logger.Debug().Msg("no groups found")
		declaration = brevity.Clean
	} else if len(friendlyGroups) > 0 && len(hostileGroups) == 0 {
		logger.Debug().Msg("friendly groups found")
		declaration = brevity.Friendly
		specificGroup = friendlyGroups[0]
	} else if len(friendlyGroups) == 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("hostile groups found")
		declaration = brevity.Hostile
		specificGroup = hostileGroups[0]
	} else if len(friendlyGroups) > 0 && len(hostileGroups) > 0 {
		logger.Debug().Msg("both friendly and hostile groups found")
		declaration = brevity.Furball
	}

	logger.Debug().Interface("declaration", declaration).Msg("responding to DECLARE request")
	c.out <- brevity.DeclareResponse{
		Callsign:    r.Callsign,
		Declaration: declaration,
		Group:       specificGroup,
	}
}
