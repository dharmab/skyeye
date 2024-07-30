package radar

import (
	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func (s *scope) FindNearbyGroups(location orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	groups := make([]brevity.Group, 0)
	itr := s.contacts.itr()
	for itr.next() {
		tf := itr.value()
		if s.isMatch(tf, coalition, filter) && geo.Distance(tf.LastKnown().Point, location) < conf.DefaultMarginRadius.Meters() {
			group := s.findGroupForAircraft(tf)
			groups = append(groups, group)
		}
	}
	return groups
}
