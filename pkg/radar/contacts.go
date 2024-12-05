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

// contactDatabase is a thread-safe trackfile contactDatabase.
type contactDatabase struct {
	lock        sync.RWMutex
	contacts    map[uint64]*trackfiles.Trackfile
	callsignIdx map[coalitions.Coalition]map[string]uint64
}

func newContactDatabase() *contactDatabase {
	db := &contactDatabase{}
	db.reset()
	return db
}

// getByCallsignAndCoalititon returns the trackfile with the lowest edit distance to the given callsign, or nil if no closely named trackfile was found.
// The second return value is true if a trackfile was found, and false otherwise.
// The callsign in the trackfile may differ from the input callsign!
func (db *contactDatabase) getByCallsignAndCoalititon(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile, bool) {
	logger := log.With().Str("callsign", callsign).Str("coalition", coalition.String()).Logger()
	db.lock.RLock()
	defer db.lock.RUnlock()

	foundCallsign := ""
	id, ok := db.callsignIdx[coalition][callsign]
	if ok {
		foundCallsign = callsign
	} else {
		keys := make([]string, 0, len(db.callsignIdx[coalition]))
		for k := range db.callsignIdx[coalition] {
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
		id = db.callsignIdx[coalition][foundCallsign]
	}
	contact, ok := db.contacts[id]
	if !ok {
		return "", nil, false
	}
	return foundCallsign, contact, true
}

// getByID returns the trackfile for the given unit ID, or nil if no trackfile was found.
// The second return value is true if a trackfile was found, and false otherwise.
func (db *contactDatabase) getByID(id uint64) (*trackfiles.Trackfile, bool) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	contact, ok := db.contacts[id]
	return contact, ok
}

// set updates the trackfile for the given trackfile's unit ID, or inserts a new trackfile if no trackfile was found.
func (db *contactDatabase) set(trackfile *trackfiles.Trackfile) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// TODO get this string munging out of here
	callsign, _, _ := strings.Cut(trackfile.Contact.Name, "|")
	callsign, ok := parser.ParsePilotCallsign(callsign)
	if !ok {
		callsign = trackfile.Contact.Name
	}
	db.callsignIdx[trackfile.Contact.Coalition][callsign] = trackfile.Contact.ID
	db.contacts[trackfile.Contact.ID] = trackfile
}

// delete removes the trackfile for the given unit ID.
func (db *contactDatabase) delete(id uint64) bool {
	db.lock.Lock()
	defer db.lock.Unlock()

	contact, ok := db.contacts[id]
	if ok {
		delete(db.callsignIdx[contact.Contact.Coalition], contact.Contact.Name)
	}
	delete(db.contacts, id)

	return ok
}

// reset removes all trackfiles from the database.
func (db *contactDatabase) reset() {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.contacts = make(map[uint64]*trackfiles.Trackfile)
	db.callsignIdx = make(map[coalitions.Coalition]map[string]uint64)
	for _, c := range []coalitions.Coalition{coalitions.Blue, coalitions.Red, coalitions.Neutrals} {
		db.callsignIdx[c] = make(map[string]uint64)
	}
}

// values iterates over all trackfiles in the database.
func (db *contactDatabase) values() iter.Seq[*trackfiles.Trackfile] {
	db.lock.RLock()
	defer db.lock.RUnlock()

	contacts := make([]*trackfiles.Trackfile, 0, len(db.contacts))
	for _, contact := range db.contacts {
		contacts = append(contacts, contact)
	}
	return slices.Values(contacts)
}
