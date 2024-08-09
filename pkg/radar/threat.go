package radar

import (
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
)

func (s *scope) Threats(coalition coalitions.Coalition) map[uint32][]brevity.Group {
	threats := make(map[uint32][]brevity.Group)
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()

		if trackfile.Contact.Coalition == coalition {
			continue
		}

		if !isValidTrack(trackfile) {
			continue
		}

		groups := s.findNearbyGroups(trackfile.LastKnown().Point, 0, 100000, 100*unit.NauticalMile, coalition, brevity.FixedWing)

		threatGroups := []brevity.Group{}
		for _, group := range groups {
			origin := trackfile.LastKnown().Point
			bearing := bearings.NewTrueBearing(unit.Angle(geo.Bearing(origin, group.point())) * unit.Degree).Magnetic(s.Declination(origin))
			_range := unit.Length(geo.Distance(origin, group.point()))
			altitude := group.Altitude()
			aspect := brevity.AspectFromAngle(bearing, group.course())
			braa := brevity.NewBRAA(bearing, _range, altitude, aspect)
			group.braa = braa
			group.bullseye = nil

			// TODO factor in aircraft threat range
			if _range < brevity.MandatoryThreatDistance || aspect == brevity.Hot {
				threatGroups = append(threatGroups, group)
			}
		}

		if len(threatGroups) != 0 {
			threats[trackfile.Contact.UnitID] = threatGroups
		}
	}
	return threats
}
