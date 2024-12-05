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

func (r *Radar) findNearbyGroups(pointOfInterest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []*group {
	circle := geo.NewBoundAroundPoint(pointOfInterest, radius.Meters())
	groups := make([]*group, 0)
	visited := make(map[uint64]struct{})
	for trackfile := range r.contacts.values() {
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
			grp := r.findGroupForAircraft(trackfile)
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

// FindNearbyGroupsWithBullseye returns all groups within the given radius of the given point of interest, within the given
// altitude block, filtered by the given coalition and contact category. Any given unit IDs are excluded from the search.
// Each group has Bullseye set relative to the point provided in SetBullseye. The groups are ordered by increasing distance
// from the point of interest.
func (r *Radar) FindNearbyGroupsWithBullseye(interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []brevity.Group {
	groups := r.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter, excludedIDs)
	result := make([]brevity.Group, 0, len(groups))
	for _, grp := range groups {
		result = append(result, grp)
	}
	return result
}

// FindNearbyGroupsWithBRAA returns all groups within the given radius of the given point of interest, within the given
// altitude block, filtered by the given coalition and contact category. Any given unit IDs are excluded from the search.
// Each group has BRAA set relative to the given origin. The groups are ordered by increasing distance from the point
// of interest.
func (r *Radar) FindNearbyGroupsWithBRAA(origin, interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory, excludedIDs []uint64) []brevity.Group {
	groups := r.findNearbyGroups(interest, minAltitude, maxAltitude, radius, coalition, filter, excludedIDs)
	result := make([]brevity.Group, 0, len(groups))
	for _, grp := range groups {
		bearing := spatial.TrueBearing(origin, grp.point()).Magnetic(r.Declination(origin))
		_range := spatial.Distance(origin, grp.point())
		aspect := brevity.AspectFromAngle(bearing, grp.course())
		grp.braa = brevity.NewBRAA(bearing, _range, grp.altitudes(), aspect)
		grp.bullseye = nil

		result = append(result, grp)
	}

	return result
}
