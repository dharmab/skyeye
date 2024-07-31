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
		tf := itr.value()
		if s.isMatch(tf, coalition, filter) && circle.Contains(tf.LastKnown().Point) {
			group := s.findGroupForAircraft(tf)
			groups = append(groups, group)
		}
	}
	return groups
}
