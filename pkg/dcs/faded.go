package dcs

import "time"

// Faded is a message that indicates an aircraft has disappeared from the unit stream.
// This can happen for many reasons including:
//
// - The aircraft has landed?
//
// - The aircraft has been destroyed.
//
// - The aircraft has been damaged and DCS considers it dead, but it may still be flying.
// This includes light damage from a simple bonk from your wingman!
//
// - A player has disconnected?
type Faded interface {
	Faded()
	// Timestamp returns the real-world clock time of the observation.
	Timestamp() time.Time
	// UnitID is the unit's in-game ID.
	UnitID() uint32
}

type faded struct {
	timestamp time.Time
	unitID    uint32
}

var _ Faded = &faded{}

func (f *faded) Faded() {}

func (f *faded) Timestamp() time.Time {
	return f.timestamp
}

func (f *faded) UnitID() uint32 {
	return f.unitID
}
