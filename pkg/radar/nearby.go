package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func (s *scope) findNearbyGroups(interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) []*group {
	circle := geo.NewBoundAroundPoint(interest, float64(radius.Meters()))
	groups := make([]*group, 0)
	visited := make(map[uint32]struct{})
	for trackfile := range s.contacts.values() {
		if _, ok := visited[trackfile.Contact.UnitID]; ok {
			continue
		}
		isMatch := s.isMatch(trackfile, coalition, filter)
		inCircle := circle.Contains(trackfile.LastKnown().Point)
		inStack := minAltitude <= trackfile.LastKnown().Altitude && trackfile.LastKnown().Altitude <= maxAltitude
		if isMatch && inCircle && inStack {
			grp := s.findGroupForAircraft(trackfile)
			for id := range grp.UnitIDs() {
				visited[uint32(id)] = struct{}{}
			}
			groups = append(groups, grp)
		}
	}

	return groups
}

func (s *scope) FindNearbyGroupsWithBullseye(interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	groups := s.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter)
	result := make([]brevity.Group, 0, len(groups))
	for _, grp := range groups {
		result = append(result, grp)
	}
	return result
}

func (s *scope) FindNearbyGroupsWithBRAA(origin, interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	groups := s.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter)
	result := make([]brevity.Group, 0, len(groups))
	for _, grp := range groups {
		bearing := spatial.TrueBearing(origin, grp.point()).Magnetic(s.Declination(origin))
		_range := spatial.Distance(origin, grp.point())
		aspect := brevity.AspectFromAngle(bearing, grp.course())
		grp.braa = brevity.NewBRAA(bearing, _range, grp.altitudes(), aspect)
		grp.bullseye = nil

		result = append(result, grp)
	}

	return result
}
