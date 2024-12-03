package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"golang.org/x/exp/slices"
)

func (s *scope) findNearbyGroups(pointOfInterest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []*group {
	circle := geo.NewBoundAroundPoint(pointOfInterest, radius.Meters())
	groups := make([]*group, 0)
	visited := make(map[uint64]struct{})
	for trackfile := range s.contacts.values() {
		if slices.Contains(excludedIDs, trackfile.Contact.ID) {
			continue
		}
		if _, ok := visited[trackfile.Contact.ID]; ok {
			continue
		}
		isMatch := isMatch(trackfile, coalition, filter)
		inCircle := circle.Contains(trackfile.LastKnown().Point)
		inStack := minAltitude <= trackfile.LastKnown().Altitude && trackfile.LastKnown().Altitude <= maxAltitude
		if isMatch && inCircle && inStack {
			grp := s.findGroupForAircraft(trackfile)
			for _, id := range grp.ObjectIDs() {
				visited[id] = struct{}{}
			}
			groups = append(groups, grp)
		}
	}

	// Sort closest to furthest
	slices.SortFunc(groups, func(a, b *group) int {
		distanceToA := spatial.Distance(pointOfInterest, a.point())
		distanceToB := spatial.Distance(pointOfInterest, b.point())
		return int(distanceToA - distanceToB)
	})

	return groups
}

func (s *scope) FindNearbyGroupsWithBullseye(interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []brevity.Group {
	groups := s.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter, excludedIDs)
	result := make([]brevity.Group, 0, len(groups))
	for _, grp := range groups {
		result = append(result, grp)
	}
	return result
}

func (s *scope) FindNearbyGroupsWithBRAA(origin, interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []brevity.Group {
	groups := s.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter, excludedIDs)
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
