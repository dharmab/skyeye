package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
)

// Threats returns a map of threat groups of the given coalition to threatened object IDs.
func (r *Radar) Threats(coalition coalitions.Coalition) map[brevity.Group][]uint64 {
	threats := make(map[*group][]uint64)
	hostileGroups := r.enumerateGroups(coalition)
	radius := 100 * unit.NauticalMile
	if r.mandatoryThreatRadius > radius {
		radius = r.mandatoryThreatRadius
	}
	for _, grp := range hostileGroups {
		friendlyGroups := r.findNearbyGroups(
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
			withinThreatRadius := r.isGroupWithinThreatRadius(grp, distance)
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
			trackfile, ok := r.contacts.getByID(threats[grp][0])
			if !ok {
				continue
			}
			declination := r.Declination(trackfile.LastKnown().Point)
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

func (r *Radar) isGroupWithinThreatRadius(grp *group, distance unit.Length) bool {
	if !grp.isArmed() {
		return false
	}
	return distance < grp.threatRadius() || distance < r.mandatoryThreatRadius
}
