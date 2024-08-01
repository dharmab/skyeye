package radar

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
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
	// FindNearbyGroups returns all groups within the given radius of the given location, filtered by the given contact category.
	// Location data is unset, since it is within radar margins of the given location.
	FindNearbyGroups(orb.Point, unit.Length, coalitions.Coalition, brevity.ContactCategory) []brevity.Group
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
	updates   <-chan sim.Updated
	fades     <-chan sim.Faded
	bullseyes <-chan orb.Point
	bullseye  orb.Point
	contacts  contactDatabase
}

func New(coalition coalitions.Coalition, bullseyes <-chan orb.Point, updates <-chan sim.Updated, fades <-chan sim.Faded) Radar {
	return &scope{
		updates:   updates,
		fades:     fades,
		bullseyes: bullseyes,
		contacts:  newContactDatabase(),
	}
}

// Run implements [Radar.Run]
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

// handleUpdate updates the database using the provided update.
func (s *scope) handleUpdate(update sim.Updated) {
	logger := log.With().
		Str("name", update.Aircraft.Name).
		Str("aircraft", update.Aircraft.ACMIName).
		Int("unitID", int(update.Aircraft.UnitID)).
		Logger()

	tf, ok := s.contacts.getByUnitID(update.Aircraft.UnitID)
	if ok {
		tf.Update(update.Frame)
		logger.Trace().Msg("updated existing trackfile")
	} else {
		tf = trackfile.NewTrackfile(update.Aircraft)
		s.contacts.set(tf)
		logger.Info().Msg("created new trackfile")
	}
}

// handleFade removed any trackfiles for the faded unit.
func (s *scope) handleFade(fade sim.Faded) {
	tf, ok := s.contacts.getByUnitID(fade.UnitID)
	if !ok {
		log.Trace().Uint32("unitID", fade.UnitID).Msg("faded trackfile not found - probably not an aircraft")
		return
	}
	logger := log.With().
		Int("unitID", int(fade.UnitID)).
		Str("name", tf.Contact.Name).
		Str("aircraft", tf.Contact.ACMIName).
		Logger()

	if !ok {
		logger.Warn().Msg("faded trackfile not found")
		return
	}
	s.contacts.delete(fade.UnitID)
	logger.Info().Msg("removed faded trackfile")
	// TODO pass fade to controller to broadcast message
}

// handleGarbageCollection removes trackfiles that have not been updated in a long time.
func (s *scope) handleGarbageCollection() {
	itr := s.contacts.itr()
	for itr.next() {
		tf := itr.value()
		logger := log.With().
			Int("unitID", int(tf.Contact.UnitID)).
			Str("name", tf.Contact.Name).
			Str("aircraft", tf.Contact.ACMIName).
			Logger()

		lastSeen, ok := s.contacts.lastUpdated(tf.Contact.UnitID)
		if !ok {
			logger.Warn().Msg("last updated time is missing")
			continue
		}
		if lastSeen.Before(time.Now().Add(-5 * time.Minute)) {
			s.contacts.delete(tf.Contact.UnitID)
			logger.Info().
				Dur("age", time.Since(lastSeen)).
				Msg("removed aged out trackfile")
		}
	}
}

// GetBullseye implements [Radar.GetBullseye]
func (s *scope) GetBullseye() orb.Point {
	return s.bullseye
}

// isValidTrack checks if the trackfile is valid. This means all of the following conditions are met:
// Last known position is not (0, 0)
// Speed is above 50 knots
// Altitude is above 10 meters above sea level
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

// isMatch checks:
// - if the trackfile is of the given coalition
// - if the trackfile is of the given contact category (or if the aircraft is not in the encyclopedia)
// - if the trackfile is valid
func (s *scope) isMatch(tf *trackfile.Trackfile, coalition coalitions.Coalition, filter brevity.ContactCategory) bool {
	if tf.Contact.Coalition != coalition {
		return false
	}
	if !isValidTrack(tf) {
		return false
	}
	data, ok := encyclopedia.GetAircraftData(tf.Contact.ACMIName)
	// If the aircraft is not in the encyclopedia, assume it matches
	matchesFilter := !ok || data.Category() == filter || filter == brevity.Aircraft
	return matchesFilter
}
