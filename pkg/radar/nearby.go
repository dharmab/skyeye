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
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			if !isValidTrack(tf) {
				continue
			}
			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category() == filter || filter == brevity.Aircraft
			if matchesFilter {
				if geo.Distance(tf.LastKnown().Point, location) < conf.DefaultMarginRadius.Meters() {
					group := s.findGroupForAircraft(tf)
					groups = append(groups, group)
				}
			}
		}
	}
	return groups
}
