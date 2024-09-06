package radar

import (
	"iter"
	"slices"
	"strings"
	"sync"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	fuzz "github.com/hbollon/go-edlib"
	"github.com/rs/zerolog/log"
)

// contactDatabase is a thread-safe trackfile database.
type contactDatabase interface {
	// getByCallsignAndCoalititon returns the trackfile with the lowest edit distance to the given callsign, or nil if no closely named trackfile was found.
	// The second return value is true if a trackfile was found, and false otherwise.
	// The callsign in the trackfile may differ from the input callsign!
	getByCallsignAndCoalititon(string, coalitions.Coalition) (string, *trackfiles.Trackfile, bool)
	// getByID returns the trackfile for the given unit ID, or nil if no trackfile was found.
	// The second return value is true if a trackfile was found, and false otherwise.
	getByID(uint64) (*trackfiles.Trackfile, bool)
	// set updates the trackfile for the given trackfile's unit ID, or inserts a new trackfile if no trackfile was found.
	set(*trackfiles.Trackfile)
	// delete removes the trackfile for the given unit ID.
	// It returns true if the trackfile was found and removed, and false otherwise.
	delete(uint64) bool
	// reset removes all trackfiles from the database.
	reset()
	// values iterates over all trackfiles in the database.
	values() iter.Seq[*trackfiles.Trackfile]
}

type database struct {
	lock        sync.RWMutex
	contacts    map[uint64]*trackfiles.Trackfile
	callsignIdx map[coalitions.Coalition]map[string]uint64
}

func newContactDatabase() contactDatabase {
	d := &database{}
	d.reset()
	return d
}

// getByCallsignAndCoalititon implements [contactDatabase.getByCallsignAndCoalititon].
func (d *database) getByCallsignAndCoalititon(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile, bool) {
	logger := log.With().Str("callsign", callsign).Str("coalition", coalition.String()).Logger()
	d.lock.RLock()
	defer d.lock.RUnlock()

	foundCallsign := ""
	id, ok := d.callsignIdx[coalition][callsign]
	if ok {
		foundCallsign = callsign
	} else {
		keys := make([]string, 0, len(d.callsignIdx[coalition]))
		for k := range d.callsignIdx[coalition] {
			keys = append(keys, k)
		}
		logger.Info().Msg("callsign not found in index, attempting fuzzy search")
		var err error
		foundCallsign, err = fuzz.FuzzySearchThreshold(callsign, keys, 0.63, fuzz.Levenshtein)
		if foundCallsign == "" || err != nil {
			logger.Warn().Err(err).Msg("callsign not found in index")
			return "", nil, false
		}
		logger.Info().Str("foundCallsign", foundCallsign).Msg("similar callsign found in index")
		id = d.callsignIdx[coalition][foundCallsign]
	}
	contact, ok := d.contacts[id]
	if !ok {
		return "", nil, false
	}
	return foundCallsign, contact, true
}

// getByID implements [contactDatabase.getByID].
func (d *database) getByID(id uint64) (*trackfiles.Trackfile, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	contact, ok := d.contacts[id]
	if !ok {
		return nil, false
	}
	return contact, true
}

// set implements [contactDatabase.set].
func (d *database) set(trackfile *trackfiles.Trackfile) {
	d.lock.Lock()
	defer d.lock.Unlock()

	// TODO get this string munging out of here
	callsign, _, _ := strings.Cut(trackfile.Contact.Name, "|")
	callsign, ok := parser.ParsePilotCallsign(callsign)
	if !ok {
		callsign = trackfile.Contact.Name
	}
	d.callsignIdx[trackfile.Contact.Coalition][callsign] = trackfile.Contact.ID
	d.contacts[trackfile.Contact.ID] = trackfile
}

// delete implements [contactDatabase.delete].
func (d *database) delete(id uint64) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	contact, ok := d.contacts[id]
	if ok {
		delete(d.callsignIdx[contact.Contact.Coalition], contact.Contact.Name)
	}
	delete(d.contacts, id)

	return ok
}

// reset implements [contactDatabase.reset].
func (d *database) reset() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.contacts = make(map[uint64]*trackfiles.Trackfile)
	d.callsignIdx = make(map[coalitions.Coalition]map[string]uint64)
	for _, c := range []coalitions.Coalition{coalitions.Blue, coalitions.Red, coalitions.Neutrals} {
		d.callsignIdx[c] = make(map[string]uint64)
	}
}

// values implements [contactDatabase.values].
func (d *database) values() iter.Seq[*trackfiles.Trackfile] {
	d.lock.RLock()
	defer d.lock.RUnlock()

	contacts := make([]*trackfiles.Trackfile, 0, len(d.contacts))
	for _, contact := range d.contacts {
		contacts = append(contacts, contact)
	}
	return slices.Values(contacts)
}
