package radar

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
)

func (s *scope) enumerateGroups(coalition coalitions.Coalition) []*group {
	visited := make(map[uint32]bool)
	groups := make([]*group, 0)
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()

		if _, ok := visited[trackfile.Contact.UnitID]; ok {
			continue
		}
		visited[trackfile.Contact.UnitID] = true

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
		for _, contact := range grp.contacts {
			visited[contact.Contact.UnitID] = true
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
	grp := newGroupUsingBullseye(bullseye)
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
	itr := s.contacts.itr()
	for itr.next() {
		other := itr.value()
		// Skip if this one is already in the group
		if slices.ContainsFunc(group.contacts, func(t *trackfiles.Trackfile) bool {
			if t == nil {
				return false
			}
			return t.Contact.UnitID == other.Contact.UnitID
		}) {
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
		isWithinSpread := geo.Distance(other.LastKnown().Point, this.LastKnown().Point) < spreadInterval.Meters()
		if !isWithinSpread {
			continue
		}

		group.contacts = append(group.contacts, other)
		s.addNearbyAircraftToGroup(other, group)
	}
}
