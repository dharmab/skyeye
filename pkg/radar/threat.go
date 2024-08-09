package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
)

func (s *scope) Threats(coalition coalitions.Coalition) map[brevity.Group][]uint32 {
	threats := make(map[*group][]uint32)
	hostileGroups := s.enumerateGroups(coalition)
	for _, grp := range hostileGroups {
		friendlyGroups := s.findNearbyGroups(
			grp.point(),
			0,
			math.MaxFloat64,
			100*unit.NauticalMile,
			coalition.Opposite(),
			brevity.Aircraft,
		)

		// Populate threats map with hostile-friendly relations that meet threat criteria.
		unitIDs := make([]uint32, 0)
		for _, friendlyGroup := range friendlyGroups {
			distance := unit.Length(geo.Distance(grp.point(), friendlyGroup.point())) * unit.Meter
			withinThreatRadius := distance < grp.threatRadius() || distance < s.mandatoryThreatRadius
			groupIsRotaryAgainstPlane := grp.category() == brevity.RotaryWing && friendlyGroup.category() == brevity.FixedWing
			if withinThreatRadius && !groupIsRotaryAgainstPlane {
				unitIDs = append(unitIDs, friendlyGroup.UnitIDs()...)
			}
		}
		if len(unitIDs) == 0 {
			continue
		}
		threats[grp] = unitIDs

		// If the hostile group only threatens a single friendly unit, use BRAA instead of Bullseye.
		if len(threats[grp]) == 1 {
			trackfile, ok := s.contacts.getByUnitID(threats[grp][0])
			if !ok {
				continue
			}
			bearing := bearings.NewTrueBearing(
				unit.Angle(
					geo.Bearing(trackfile.LastKnown().Point, grp.point()),
				) * unit.Degree,
			).Magnetic(s.Declination(trackfile.LastKnown().Point))
			_range := unit.Length(geo.Distance(trackfile.LastKnown().Point, grp.point()))
			aspect := brevity.AspectFromAngle(bearing, grp.course())
			grp.braa = brevity.NewBRAA(bearing, _range, grp.Altitude(), aspect)
			grp.bullseye = nil
		}
	}

	result := make(map[brevity.Group][]uint32)
	for grp, unitIDs := range threats {
		result[grp] = unitIDs
	}
	return result
}
