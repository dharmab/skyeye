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

// Merges returns a map of hostile groups of the given coalition to friendly trackfiles.
func (r *Radar) Merges(coalition coalitions.Coalition) map[brevity.Group][]*trackfiles.Trackfile {
	visited := make(map[uint64]struct{})
	merges := make(map[brevity.Group][]*trackfiles.Trackfile)
	bullseye := r.Bullseye(coalition)
	for contact := range r.contacts.values() {
		if _, ok := visited[contact.Contact.ID]; ok {
			continue
		}

		if contact.Contact.Coalition != coalition.Opposite() {
			continue
		}

		if !isValidTrack(contact) {
			continue
		}

		// Exclude hostile helicopters
		if data, ok := encyclopedia.GetAircraftData(contact.Contact.ACMIName); ok && data.Category() == brevity.RotaryWing {
			continue
		}

		grp := r.findGroupForAircraft(contact)
		mergedWith := make(map[uint64]*trackfiles.Trackfile)
		for _, contact := range grp.contacts {
			visited[contact.Contact.ID] = struct{}{}
			for _, trackfile := range r.mergesForContact(contact) {
				mergedWith[trackfile.Contact.ID] = trackfile
			}
		}
		if len(mergedWith) == 0 {
			continue
		}
		grp.isThreat = true
		grp.bullseye = &bullseye
		grp.SetDeclaration(brevity.Furball)

		merges[grp] = slices.Collect(maps.Values(mergedWith))
	}
	return merges
}

// mergesForContact returns the opposing trackfiles that the given trackfile is merged with.
func (r *Radar) mergesForContact(trackfile *trackfiles.Trackfile) []*trackfiles.Trackfile {
	mergedWith := make([]*trackfiles.Trackfile, 0)
	if trackfile.IsLastKnownPointZero() {
		return mergedWith
	}
	for other := range r.contacts.values() {
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
