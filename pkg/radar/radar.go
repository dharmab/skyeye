package radar

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Radar interface {
	// Run consumes updates from the simulation channels until the context is cancelled.
	Run(context.Context)
	// FindCallsign returns the trackfile for the given callsign, or nil if no trackfile was found.
	FindCallsign(string) *trackfile.Trackfile
	// FindUnit returns the trackfile for the given unit ID, or nil if no trackfile was found.
	FindUnit(uint32) *trackfile.Trackfile
	// GetBullseye returns the bullseye for the configured coalition.
	GetBullseye() orb.Point
	// GetPicture returns a picture of the radar scope around the given location, within the given radius, filtered by the given coalition and contact category.
	// The first return value is the total number of groups, and the second is a slice of up to to 3 high priority groups.
	GetPicture(orb.Point, unit.Length, coalitions.Coalition, brevity.ContactCategory) (int, []brevity.Group)
	// FindNearbyGroups returns all groups within 3 nautical miles of the given location, filtered by the given contact category.
	// Location data is unset, since it is within radar margins of the given location.
	FindNearbyGroups(orb.Point, coalitions.Coalition, brevity.ContactCategory) []brevity.Group
	// FindNearestGroupWithBRAA returns the nearest group to the given location, with BRAA location embedded in the Group.
	// The given point is the location to search from.
	// The given coalition is the coalition to search for.
	// The given filter is the contact category to filter by.
	// Returns the nearest group to the given location which matches the given coalition and filter, with BRAA relative to the given location. Returns nil if no group was found.
	FindNearestGroupWithBRAA(orb.Point, coalitions.Coalition, brevity.ContactCategory) brevity.Group
	// FindNearestGroupWithBullseye returns the nearest group to the given location, with Bullseye location embedded in the Group.
	// The given point is the location to search from.
	// The given coalition is the coalition to search for.
	// The given filter is the contact category to filter by.
	// Returns the nearest group to the given location which matches the given coalition and filter, with Bullseye location. Returns nil if no group was found.
	FindNearestGroupWithBullseye(orb.Point, coalitions.Coalition, brevity.ContactCategory) brevity.Group
	// FindNearestGroupInCone returns the nearest group to the given location along the given bearing, ± the given angle, with BRAA relative to the given location. Returns nil if no group was found.
	FindNearestGroupInCone(orb.Point, unit.Angle, unit.Angle, coalitions.Coalition, brevity.ContactCategory) brevity.Group
}

var _ Radar = &scope{}

type scope struct {
	updates      <-chan sim.Updated
	fades        <-chan sim.Faded
	bullseyes    <-chan orb.Point
	lock         sync.RWMutex
	callsignIdx  map[string]uint32
	contacts     map[uint32]*trackfile.Trackfile
	bullseye     orb.Point
	aircraftData map[string]encyclopedia.Aircraft
}

func New(coalition coalitions.Coalition, bullseyes <-chan orb.Point, updates <-chan sim.Updated, fades <-chan sim.Faded) Radar {
	e := encyclopedia.New()

	return &scope{
		updates:      updates,
		fades:        fades,
		callsignIdx:  make(map[string]uint32),
		contacts:     make(map[uint32]*trackfile.Trackfile),
		bullseyes:    bullseyes,
		aircraftData: e.Aircraft(),
		lock:         sync.RWMutex{},
	}
}

func (s *scope) Run(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case update := <-s.updates:
			s.handleUpdate(update)
		case fade := <-s.fades:
			s.handleFade(fade)
		case bullseye := <-s.bullseyes:
			s.bullseye = bullseye
		case <-ticker.C:
			s.handleGarbageCollection()
		case <-ctx.Done():
			return
		}
	}
}

func (s *scope) handleUpdate(update sim.Updated) {
	callsign, _, _ := strings.Cut(update.Aircraft.Name, "|")
	// replace digits and spaces with digit followed by a single space
	callsign, ok := parser.ParseCallsign(callsign)
	if !ok {
		callsign = update.Aircraft.Name
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	unitID, ok := s.callsignIdx[callsign]
	logger := log.With().
		Str("callsign", callsign).
		Str("aircraft", update.Aircraft.ACMIName).
		Str("name", update.Aircraft.ACMIName).
		Int("unitID", int(update.Aircraft.UnitID)).
		Logger()

	if ok && unitID != update.Aircraft.UnitID {
		logger.Warn().Int("otherUnitID", int(unitID)).Msg("callsigns conflict")
		s.contacts[update.Aircraft.UnitID] = trackfile.NewTrackfile(update.Aircraft)
		logger.Info().Msg("overwrote trackfile")
	}

	if !ok {
		s.contacts[update.Aircraft.UnitID] = trackfile.NewTrackfile(update.Aircraft)
		age := time.Since(update.Frame.Timestamp)
		logger.Info().Dur("age", age).Msg("new trackfile")
	}
	contact, ok := s.contacts[update.Aircraft.UnitID]
	if ok {
		contact.Update(update.Frame)
		s.callsignIdx[callsign] = update.Aircraft.UnitID
		logger.Trace().Msg("updated trackfile")
	}
}

func (s *scope) handleFade(fade sim.Faded) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.removeTrack(fade.UnitID, "removed faded trackfile")
	// after some time, send faded message to controller?
}

func (s *scope) removeTrack(unitID uint32, reason string) {
	tf, ok := s.contacts[unitID]
	if ok {
		logger := log.With().
			Int("unitID", int(unitID)).
			Str("name", tf.Contact.Name).
			Str("aircraft", tf.Contact.ACMIName).
			Dur("age", time.Since(tf.LastKnown().Timestamp)).
			Logger()

		delete(s.contacts, unitID)
		for callsign, i := range s.callsignIdx {
			if i == unitID {
				delete(s.callsignIdx, callsign)
			}
			break
		}
		logger.Info().Msg(reason)
	}
}

func (s *scope) handleGarbageCollection() {
	s.lock.Lock()
	defer s.lock.Unlock()
	for unitID, tf := range s.contacts {
		if tf.LastKnown().Timestamp.Before(time.Now().Add(-30 * time.Second)) {
			s.removeTrack(unitID, "removed aged out trackfile")
		}
	}
}

func (s *scope) GetBullseye() orb.Point {
	return s.bullseye
}

func isValidTrack(tf *trackfile.Trackfile) bool {
	point := tf.LastKnown().Point
	isValidLongitude := point.Lon() != 0
	isValidLatitude := point.Lat() != 0
	isValidPosition := isValidLongitude && isValidLatitude
	isAboveSpeedFilter := tf.Speed() > 50*unit.Knot
	isAboveAltitudeFilter := tf.LastKnown().Altitude > 10*unit.Meter
	isValid := isValidPosition && isAboveSpeedFilter && isAboveAltitudeFilter
	log.Debug().
		Str("aircraft", tf.Contact.ACMIName).
		Int("unitID", int(tf.Contact.UnitID)).
		Str("callsign", tf.Contact.Name).
		Bool("isValid", isValid).
		Bool("isValidLongitude", isValidLongitude).
		Bool("isValidLatitude", isValidLatitude).
		Bool("isValidPosition", isValidPosition).
		Bool("isAboveSpeedFilter", isAboveSpeedFilter).
		Bool("isAboveAltitudeFilter", isAboveAltitudeFilter).
		Msg("checking track validity")
	return isValid
}
