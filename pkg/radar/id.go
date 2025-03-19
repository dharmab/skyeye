package radar

import (
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

// FindCallsign returns the trackfile on the given coalition that mosty closely matches the given callsign,
// or nil if no closely matching trackfile was found.
// The first return value is the callsign of the trackfile, and the second is the trackfile itself.
// The returned callsign may differ from the input callsign!
func (r *Radar) FindCallsign(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile) {
	foundCallsign, tf, ok := r.contacts.getByCallsignAndCoalititon(callsign, coalition)
	if !ok {
		return callsign, nil
	}
	return foundCallsign, tf
}

// FindUnit returns the trackfile for the given unit ID, or nil if no trackfile was found.
func (r *Radar) FindUnit(id uint64) *trackfiles.Trackfile {
	trackfile, ok := r.contacts.getByID(id)
	if !ok {
		return nil
	}
	return trackfile
}

// FindByCoalition returns all trackfiles on the given coalition.
func (r *Radar) FindByCoalition(coalition coalitions.Coalition) []*trackfiles.Trackfile {
	contacts := []*trackfiles.Trackfile{}
	for trackfile := range r.contacts.values() {
		if trackfile.Contact.Coalition == coalition {
			contacts = append(contacts, trackfile)
		}
	}
	return contacts
}
