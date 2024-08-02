package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func (s *scope) FindNearbyGroups(location orb.Point, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	circle := geo.NewBoundAroundPoint(location, float64(radius.Meters()))
	groups := make([]brevity.Group, 0)
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		if s.isMatch(trackfile, coalition, filter) && circle.Contains(trackfile.LastKnown().Point) {
			group := s.findGroupForAircraft(trackfile)
			groups = append(groups, group)
		}
	}
	return groups
}
