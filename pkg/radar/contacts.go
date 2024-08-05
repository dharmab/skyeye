package radar

import (
	"strings"
	"sync"
	"time"

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
	// lastUpdated returns the last time a trackfile was updated, using the real time timestamp.
	lastUpdated(uint32) (time.Time, bool)
	// set updates the trackfile for the given trackfile's unit ID, or inserts a new trackfile if no trackfile was found.
	set(*trackfiles.Trackfile)
	// delete removes the trackfile for the given unit ID.
	// It returns true if the trackfile was found and removed, and false otherwise.
	delete(uint32) bool
	// itr returns an iterator over the database.
	itr() databaseIterator
}

// databaseIterator iterates over the contents of a contactDatabase.
type databaseIterator interface {
	// next advances the iterator to the next trackfile in the database.
	// It returns false when the iterator has passed the last trackfile.
	next() bool
	// reset the iterator to the beginning.
	reset()
	// value returns the trackfile at the current position of the iterator.
	// It should only be called after Next returns true.
	value() *trackfiles.Trackfile
}

type database struct {
	lock           sync.RWMutex
	contacts       map[uint32]*trackfiles.Trackfile
	callsignIdx    map[coalitions.Coalition]map[string]uint32
	lastUpdatedIdx map[uint32]time.Time
}

func newContactDatabase() contactDatabase {
	callsignIdx := make(map[coalitions.Coalition]map[string]uint32)
	for _, c := range []coalitions.Coalition{coalitions.Blue, coalitions.Red, coalitions.Neutrals} {
		callsignIdx[c] = make(map[string]uint32)
	}
	return &database{
		contacts:       make(map[uint32]*trackfiles.Trackfile),
		callsignIdx:    callsignIdx,
		lastUpdatedIdx: make(map[uint32]time.Time),
	}
}

// getByCallsignAndCoalititon implements [contactDatabase.getByCallsignAndCoalititon].
func (d *database) getByCallsignAndCoalititon(callsign string, coalition coalitions.Coalition) (string, *trackfiles.Trackfile, bool) {
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
		log.Info().Str("callsign", callsign).Msg("callsign not found in index, attempting fuzzy search")
		var err error
		foundCallsign, err = fuzz.FuzzySearchThreshold(callsign, keys, 0.63, fuzz.Levenshtein)
		if foundCallsign == "" || err != nil {
			log.Warn().Err(err).Str("callsign", callsign).Msg("callsign not found in index")
			return "", nil, false
		}
		log.Info().Str("callsign", callsign).Str("foundCallsign", foundCallsign).Msg("similar callsign found in index")
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
	d.lastUpdatedIdx[trackfile.Contact.UnitID] = time.Now()
}

// lastUpdated implements [contactDatabase.lastUpdated].
func (d *database) lastUpdated(unitId uint32) (time.Time, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	lastUpdated, ok := d.lastUpdatedIdx[unitId]
	return lastUpdated, ok
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
	delete(d.lastUpdatedIdx, unitId)

	return ok
}

// itr implements [contactDatabase.itr].
func (d *database) itr() databaseIterator {
	d.lock.RLock()
	defer d.lock.RUnlock()

	// Iterate over a copy, for thread safety
	unitIds := make([]uint32, 0, len(d.contacts))
	copy := make(map[uint32]*trackfiles.Trackfile)
	for unitId := range d.contacts {
		unitIds = append(unitIds, unitId)
		copy[unitId] = d.contacts[unitId]
	}

	return newDatabaseIterator(unitIds, func(id uint32) (*trackfiles.Trackfile, bool) {
		contact, ok := copy[id]
		return contact, ok
	})
}

type iterator struct {
	cursor  int
	unitIds []uint32
	getFn   func(uint32) (*trackfiles.Trackfile, bool)
}

func newDatabaseIterator(unitIds []uint32, getFn func(uint32) (*trackfiles.Trackfile, bool)) databaseIterator {
	return &iterator{
		cursor:  -1,
		unitIds: unitIds,
		getFn:   getFn,
	}
}

// next implements [iterator.next].
func (i *iterator) next() bool {
	i.cursor++
	return i.cursor < len(i.unitIds)
}

// reset implements [iterator.reset].
func (i *iterator) reset() {
	i.cursor = -1
}

// value implements [iterator.value].
func (i *iterator) value() *trackfiles.Trackfile {
	contact, ok := i.getFn(i.unitIds[i.cursor])
	if !ok {
		return nil
	}
	return contact
}
