package radar

import (
	"maps"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

// Merges returns a map of fixed-wing groups on the opposing coalition to the contacts on the given coalition that they are merged with.
func (s *scope) Merges(coalition coalitions.Coalition) map[brevity.Group][]*trackfiles.Trackfile {
	visited := make(map[uint64]struct{})
	merges := make(map[brevity.Group][]*trackfiles.Trackfile)
	bullseye := s.Bullseye(coalition)
	for contact := range s.contacts.values() {
		if _, ok := visited[contact.Contact.ID]; ok {
			continue
		}

		if contact.Contact.Coalition != coalition.Opposite() {
			continue
		}

		if contact.IsLastKnownPointZero() {
			continue
		}

		// Exclude hostile helicopters
		if data, ok := encyclopedia.GetAircraftData(contact.Contact.ACMIName); ok && data.Category() == brevity.RotaryWing {
			continue
		}

		grp := s.findGroupForAircraft(contact)
		mergedWith := make(map[uint64]*trackfiles.Trackfile)
		for _, contact := range grp.contacts {
			visited[contact.Contact.ID] = struct{}{}
			for _, trackfile := range s.mergesForContact(contact) {
				mergedWith[trackfile.Contact.ID] = trackfile
			}
		}
		if len(mergedWith) == 0 {
			continue
		}
		grp.isThreat = true
		grp.bullseye = &bullseye
		grp.declaration = brevity.Furball

		merges[grp] = slices.Collect(maps.Values(mergedWith))
	}
	return merges
}

// mergesForContact returns the opposing trackfiles that the given trackfile is merged with.
func (s *scope) mergesForContact(trackfile *trackfiles.Trackfile) []*trackfiles.Trackfile {
	mergedWith := make([]*trackfiles.Trackfile, 0)
	if trackfile.IsLastKnownPointZero() {
		return mergedWith
	}
	for other := range s.contacts.values() {
		if trackfile.Contact.Coalition == other.Contact.Coalition {
			continue
		}
		if other.IsLastKnownPointZero() {
			continue
		}
		distance := spatial.Distance(trackfile.LastKnown().Point, other.LastKnown().Point)
		if distance < brevity.MergeExitDistance {
			mergedWith = append(mergedWith, other)
		}
	}
	return mergedWith
}
