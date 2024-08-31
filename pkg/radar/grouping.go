package radar

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
)

func (s *scope) enumerateGroups(coalition coalitions.Coalition) []*group {
	visited := make(map[uint64]struct{})
	groups := make([]*group, 0)
	for trackfile := range s.contacts.values() {
		if _, ok := visited[trackfile.Contact.ID]; ok {
			continue
		}
		visited[trackfile.Contact.ID] = struct{}{}

		if trackfile.Contact.Coalition != coalition {
			continue
		}

		if !isValidTrack(trackfile) {
			continue
		}

		grp := s.findGroupForAircraft(trackfile)
		if grp == nil {
			continue
		}
		for _, id := range grp.ObjectIDs() {
			visited[id] = struct{}{}
		}
		groups = append(groups, grp)
	}

	return groups
}

// findGroupForAircraft creates a new group for the given trackfile and adds all nearby aircraft which can be considered part of the group.
func (s *scope) findGroupForAircraft(trackfile *trackfiles.Trackfile) *group {
	if trackfile == nil {
		return nil
	}
	bullseye := s.Bullseye(trackfile.Contact.Coalition)
	grp := &group{
		bullseye:    &bullseye,
		contacts:    make([]*trackfiles.Trackfile, 0),
		declaration: brevity.Unable,
	}
	grp.contacts = append(grp.contacts, trackfile)
	s.addNearbyAircraftToGroup(trackfile, grp)
	return grp
}

// addNearbyAircraftToGroup recursively adds all nearby aircraft which:
//   - are of the same coalition
//   - are within 5 nautical miles in 2D distance of each other
//   - have similar tags
//
// The spread is increased from the ATP numbers beacause the DCS AI isn't amazing at holding formation.
// We allow mixed platform groups because these are fairly common in DCS.
func (s *scope) addNearbyAircraftToGroup(this *trackfiles.Trackfile, group *group) {
	var tag encyclopedia.AircraftTag
	thisData, ok := encyclopedia.GetAircraftData(this.Contact.ACMIName)
	if ok {
		if thisData.HasTag(encyclopedia.Fighter) {
			tag = encyclopedia.Fighter
		} else if thisData.HasTag(encyclopedia.Attack) {
			tag = encyclopedia.Attack
		} else if thisData.HasTag(encyclopedia.Unarmed) {
			tag = encyclopedia.Unarmed
		}
	}
	spreadInterval := 5 * unit.NauticalMile
	for other := range s.contacts.values() {
		// Skip if this one is already in the group
		if slices.Contains(group.ObjectIDs(), other.Contact.ID) {
			continue
		}

		// Check coalition, categoty, and filters
		if !s.isMatch(other, this.Contact.Coalition, group.category()) {
			continue
		}

		// Check tag similarity
		subCategories := []encyclopedia.AircraftTag{encyclopedia.Fighter, encyclopedia.Attack, encyclopedia.Unarmed}
		if slices.Contains(subCategories, tag) {
			data, ok := encyclopedia.GetAircraftData(other.Contact.ACMIName)
			if !ok {
				continue
			}
			if !data.HasTag(tag) {
				continue
			}
		}

		// Check spread distance
		distance := spatial.Distance(other.LastKnown().Point, this.LastKnown().Point)
		isWithinSpread := distance < spreadInterval
		if !isWithinSpread {
			continue
		}

		group.contacts = append(group.contacts, other)
		s.addNearbyAircraftToGroup(other, group)
	}
}
