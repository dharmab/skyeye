package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func (s *scope) FindNearbyGroups(origin, interest orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	circle := geo.NewBoundAroundPoint(interest, float64(radius.Meters()))
	groups := make([]brevity.Group, 0)
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		isMatch := s.isMatch(trackfile, coalition, filter)
		inCircle := circle.Contains(trackfile.LastKnown().Point)
		inStack := minAltitude <= trackfile.LastKnown().Altitude && trackfile.LastKnown().Altitude <= maxAltitude
		if isMatch && inCircle && inStack {
			grp := s.findGroupForAircraft(trackfile)

			bearing := bearings.NewTrueBearing(
				unit.Angle(
					geo.Bearing(origin, grp.point()),
				) * unit.Degree,
			).Magnetic(s.Declination(origin))
			_range := unit.Length(math.Abs(geo.Distance(origin, grp.point())))
			altitude := grp.Altitude()
			aspect := brevity.AspectFromAngle(bearing, grp.course())
			grp.braa = brevity.NewBRAA(bearing, _range, altitude, aspect)
			grp.bullseye = nil

			groups = append(groups, grp)
		}
	}

	return groups
}
