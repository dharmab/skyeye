package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/paulmach/orb/geo"
)

// HandleDeclare implements Controller.HandleDeclare.
func (c *controller) HandleDeclare(r *brevity.DeclareRequest) {
	bullseye := c.scope.GetBullseye(c.coalition).Point
	location := geo.PointAtBearingAndDistance(
		bullseye,
		r.Location.Bearing().Degrees(),
		r.Location.Distance().Meters(),
	)
	friendlyGroups := c.scope.FindNearbyGroups(location, c.coalition, brevity.Airplanes)
	hostileGroups := c.scope.FindNearbyGroups(location, c.hostileCoalition(), brevity.Airplanes)

	declaration := brevity.Unable
	var specificGroup brevity.Group
	if len(friendlyGroups)+len(hostileGroups) == 0 {
		declaration = brevity.Clean
	} else if len(friendlyGroups) > 0 && len(hostileGroups) == 0 {
		declaration = brevity.Friendly
		specificGroup = friendlyGroups[0]
	} else if len(friendlyGroups) == 0 && len(hostileGroups) > 0 {
		declaration = brevity.Hostile
		specificGroup = hostileGroups[0]
	} else if len(friendlyGroups) > 0 && len(hostileGroups) > 0 {
		declaration = brevity.Furball
	}

	c.out <- brevity.DeclareResponse{
		Callsign:    r.Callsign,
		Declaration: declaration,
		Group:       specificGroup,
	}
}
