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
	// getByUnitID returns the trackfile for the given unit ID, or nil if no trackfile was found.
	// The second return value is true if a trackfile was found, and false otherwise.
	getByUnitID(uint32) (*trackfiles.Trackfile, bool)
	// set updates the trackfile for the given trackfile's unit ID, or inserts a new trackfile if no trackfile was found.
	set(*trackfiles.Trackfile)
	// delete removes the trackfile for the given unit ID.
	// It returns true if the trackfile was found and removed, and false otherwise.
	delete(uint32) bool
	// values iterates over all trackfiles in the database.
	values() iter.Seq[*trackfiles.Trackfile]
}

type database struct {
	lock        sync.RWMutex
	contacts    map[uint32]*trackfiles.Trackfile
	callsignIdx map[coalitions.Coalition]map[string]uint32
}

func newContactDatabase() contactDatabase {
	callsignIdx := make(map[coalitions.Coalition]map[string]uint32)
	for _, c := range []coalitions.Coalition{coalitions.Blue, coalitions.Red, coalitions.Neutrals} {
		callsignIdx[c] = make(map[string]uint32)
	}
	return &database{
		contacts:    make(map[uint32]*trackfiles.Trackfile),
		callsignIdx: callsignIdx,
	}
}

// getByCallsignAndCoalititon implements [contactDatabase.getByCallsignAndCoalititon].
func (d *database) getByCallsignAndCoalititon(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile, bool) {
	logger := log.With().Str("callsign", callsign).Str("coalition", coalition.String()).Logger()
	d.lock.RLock()
	defer d.lock.RUnlock()

	foundCallsign := ""
	unitId, ok := d.callsignIdx[coalition][callsign]
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
		unitId = d.callsignIdx[coalition][foundCallsign]
	}
	contact, ok := d.contacts[unitId]
	if !ok {
		return "", nil, false
	}
	return foundCallsign, contact, true
}

// getByUnitID implements [contactDatabase.getByUnitID].
func (d *database) getByUnitID(unitId uint32) (*trackfiles.Trackfile, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	contact, ok := d.contacts[unitId]
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
	d.callsignIdx[trackfile.Contact.Coalition][callsign] = trackfile.Contact.UnitID
	d.contacts[trackfile.Contact.UnitID] = trackfile
}

// delete implements [contactDatabase.delete].
func (d *database) delete(unitId uint32) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	contact, ok := d.contacts[unitId]
	if ok {
		delete(d.callsignIdx[contact.Contact.Coalition], contact.Contact.Name)
	}
	delete(d.contacts, unitId)

	return ok
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
