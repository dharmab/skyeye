package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
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
			distance := spatial.Distance(grp.point(), friendlyGroup.point())
			withinThreatRadius := distance < grp.threatRadius() || distance < s.mandatoryThreatRadius
			hostileIsHelo := grp.category() == brevity.RotaryWing
			friendlyIsPlane := friendlyGroup.category() == brevity.FixedWing
			heloVersusPlane := hostileIsHelo && friendlyIsPlane
			if withinThreatRadius && !heloVersusPlane {
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
			declination := s.Declination(trackfile.LastKnown().Point)
			bearing := spatial.TrueBearing(trackfile.LastKnown().Point, grp.point()).Magnetic(declination)
			_range := spatial.Distance(trackfile.LastKnown().Point, grp.point())
			aspect := brevity.AspectFromAngle(bearing, grp.course())
			grp.braa = brevity.NewBRAA(bearing, _range, grp.altitudes(), aspect)
			grp.bullseye = nil
		}
	}

	result := make(map[brevity.Group][]uint32)
	for grp, unitIDs := range threats {
		result[grp] = unitIDs
	}
	return result
}
