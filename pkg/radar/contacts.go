package radar

import (
	"strings"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

// contactDatabase is a thread-safe trackfile database.
type contactDatabase interface {
	// getByCallsign returns the trackfile for the given callsign, or nil if no trackfile was found.
	// The second return value is true if a trackfile was found, and false otherwise.
	getByCallsign(string) (*trackfiles.Trackfile, bool)
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
	callsignIdx    map[string]uint32
	lastUpdatedIdx map[uint32]time.Time
}

func newContactDatabase() contactDatabase {
	return &database{
		contacts:       make(map[uint32]*trackfiles.Trackfile),
		callsignIdx:    make(map[string]uint32),
		lastUpdatedIdx: make(map[uint32]time.Time),
	}
}

// getByCallsign implements [contactDatabase.getByCallsign].
func (d *database) getByCallsign(callsign string) (*trackfiles.Trackfile, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	unitId, ok := d.callsignIdx[callsign]
	if !ok {
		return nil, false
	}
	contact, ok := d.contacts[unitId]
	if !ok {
		return nil, false
	}
	return contact, true
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
func (d *database) set(tf *trackfiles.Trackfile) {
	d.lock.Lock()
	defer d.lock.Unlock()

	// TODO get this string munging out of here
	callsign, _, _ := strings.Cut(tf.Contact.Name, "|")
	callsign, ok := parser.ParseCallsign(callsign)
	if !ok {
		callsign = tf.Contact.Name
	}
	d.callsignIdx[callsign] = tf.Contact.UnitID
	d.contacts[tf.Contact.UnitID] = tf
	d.lastUpdatedIdx[tf.Contact.UnitID] = time.Now()
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
		delete(d.callsignIdx, contact.Contact.Name)
	}
	delete(d.contacts, unitId)
	delete(d.lastUpdatedIdx, unitId)

	return ok
}

// itr implements [contactDatabase.itr].
func (d *database) itr() databaseIterator {
	d.lock.RLock()
	defer d.lock.RUnlock()

	unitIds := make([]uint32, 0, len(d.contacts))
	for unitId := range d.contacts {
		unitIds = append(unitIds, unitId)
	}

	return newDatabaseIterator(unitIds, d.getByUnitID)
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
