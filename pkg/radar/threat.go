package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
)

func (s *scope) Threats(coalition coalitions.Coalition) map[brevity.Group][]uint64 {
	threats := make(map[*group][]uint64)
	hostileGroups := s.enumerateGroups(coalition)
	radius := 100 * unit.NauticalMile
	if s.mandatoryThreatRadius > radius {
		radius = s.mandatoryThreatRadius
	}
	for _, grp := range hostileGroups {
		friendlyGroups := s.findNearbyGroups(
			grp.point(),
			0,
			math.MaxFloat64,
			radius,
			coalition.Opposite(),
			brevity.Aircraft,
			[]uint64{},
		)

		// Populate threats map with hostile-friendly relations that meet threat criteria.
		ids := make([]uint64, 0)
		for _, friendlyGroup := range friendlyGroups {
			distance := spatial.Distance(grp.point(), friendlyGroup.point())
			withinThreatRadius := s.isGroupWithinThreatRadius(grp, distance)
			hostileIsHelo := grp.category() == brevity.RotaryWing
			friendlyIsPlane := friendlyGroup.category() == brevity.FixedWing
			heloVersusPlane := hostileIsHelo && friendlyIsPlane
			if withinThreatRadius && !heloVersusPlane {
				ids = append(ids, friendlyGroup.ObjectIDs()...)
			}
		}
		if len(ids) == 0 {
			continue
		}
		threats[grp] = ids

		// If the hostile group only threatens a single friendly unit, use BRAA instead of Bullseye.
		if len(threats[grp]) == 1 {
			trackfile, ok := s.contacts.getByID(threats[grp][0])
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

	result := make(map[brevity.Group][]uint64)
	for grp, ids := range threats {
		result[grp] = ids
	}
	return result
}

func (s *scope) isGroupWithinThreatRadius(grp *group, distance unit.Length) bool {
	return distance < grp.threatRadius() || distance < s.mandatoryThreatRadius
}
