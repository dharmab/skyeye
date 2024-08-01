package radar

import (
	"math"
	"slices"

	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// findGroupForAircraft creates a new group for the given trackfile and adds all nearby aircraft which can be considered part of the group.
func (s *scope) findGroupForAircraft(tf *trackfiles.Trackfile) *group {
	if tf == nil {
		return nil
	}
	group := newGroupUsingBullseye(s.bullseye)
	group.contacts = append(group.contacts, tf)
	s.addNearbyAircraftToGroup(tf, group)
	return group
}

// addNearbyAircraftToGroup recursively adds all nearby aircraft which:
//
// - are of the same coalition
//
// - are within 3 nautical miles in 2D distance of each other
//
// - are within 3000 feet in altitude of each other
//
// These are tripled from the ATP numbers beacause the DCS AI isn't amazing at holding formation.
// We allow mixed platform groups because these are fairly common in DCS.
func (s *scope) addNearbyAircraftToGroup(this *trackfiles.Trackfile, group *group) {
	spreadInterval := unit.Length(3) * unit.NauticalMile
	stackInterval := unit.Length(3000) * unit.Foot
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

		if !s.isMatch(other, this.Contact.Coalition, group.category()) {
			continue
		}

		isWithinSpread := geo.Distance(other.LastKnown().Point, this.LastKnown().Point) < spreadInterval.Meters()
		isWithinStack := math.Abs(other.LastKnown().Altitude.Feet()-this.LastKnown().Altitude.Feet()) < stackInterval.Feet()
		log.Debug().
			Any("initialContact", this.Contact).
			Any("contact", other.Contact).
			Int("unitID", int(other.Contact.UnitID)).
			Bool("isWithinSpread", isWithinSpread).
			Bool("isWithinStack", isWithinStack).
			Msg("checking if contact is within group")
		if isWithinSpread && isWithinStack {
			group.contacts = append(group.contacts, other)
			s.addNearbyAircraftToGroup(other, group)
		}
	}
}
