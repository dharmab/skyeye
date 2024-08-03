// package radar implements mid-level logic for Ground-Controlled Interception (GCI)
package radar

import (
	"context"
	"math/rand/v2"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

// Radar consumes updates from the simulation, keeps track of each aircraft as a trackfile, and provides functions to collect the aircraft into groups.
type Radar interface {
	// SetBullseye updates the bullseye point. The bullseye point is the reference point for polar
	// coordinates provided in [Group.Bullseye].
	SetBullseye(orb.Point)
	// Bullseye returns the bullseye point.
	GetBullseye() orb.Point
	// SetMissionTime updates the mission time. The mission time is used for computing magnetic declination.
	SetMissionTime(time.Time)
	// Declination returns the magnetic declination at the given point, at the time provided in SetMissionTime.
	Declination(orb.Point) unit.Angle
	// Run consumes updates from the simulation channels until the context is cancelled.
	Run(context.Context)
	// FindCallsign returns the trackfile that mosty closely matches the given callsign, or nil if no closely matching trackfile was found.
	// The first return value is the callsign of the trackfile, and the second is the trackfile itself.
	// The returned callsign may differ from the input callsign!
	FindCallsign(string) (string, *trackfiles.Trackfile)
	// FindUnit returns the trackfile for the given unit ID, or nil if no trackfile was found.
	FindUnit(uint32) *trackfiles.Trackfile
	// GetPicture returns a picture of the radar scope anchored at the given origin, within the given radius,
	// filtered by the given coalition and contact category. The first return value is the total number of groups
	// and the second is a slice of up to to 3 high priority groups. Each group has Bullseye set relative to the
	// the point provided in SetBullseye.
	GetPicture(
		origin orb.Point,
		radius unit.Length,
		coalition coalitions.Coalition,
		category brevity.ContactCategory,
	) (int, []brevity.Group)
	// FindNearbyGroups returns all groups within the given radius of the given point of interest, within the given
	// altitude block, filtered by the given coalition and contact category. Each group has BRAA set relative to the
	// given origin.
	FindNearbyGroups(
		origin,
		pointOfInterest orb.Point,
		minAltitude,
		maxAltitude,
		radius unit.Length,
		coalition coalitions.Coalition,
		category brevity.ContactCategory,
	) []brevity.Group
	// FindNearestGroupWithBRAA returns the nearest group to the given origin (up to the given radius), within the
	// given altitude block, filtered by the given coalition and contact category. The group has BRAA set relative to
	// the given origin. Returns nil if no group was found.
	FindNearestGroupWithBRAA(
		origin orb.Point,
		minAltitude,
		maxAltitude,
		radius unit.Length,
		coalition coalitions.Coalition,
		category brevity.ContactCategory,
	) brevity.Group
	// FindNearestGroupWithBullseye returns the nearest group to the given point of interest (up to the given radius),
	// within the given altitude block, filtered by the given coalition and contact category. The group has Bullseye
	// set relative to the point provided in SetBullseye. Returns nil if no group was found.
	FindNearestGroupWithBullseye(
		pointOfIntest orb.Point,
		minAltitude,
		maxAltitude,
		radius unit.Length,
		coalition coalitions.Coalition,
		category brevity.ContactCategory,
	) brevity.Group
	// FindNearestGroupInSector returns the nearest group to the given origin (up to the given distance), within a 2D
	// circular sector defined by the given origin ,radius, bearing and arc, within the given altitude block, filtered
	// by the given coalition and contact category. The group has BRAA set relative to the given origin. Returns nil if
	// no group was found.
	FindNearestGroupInSector(
		origin orb.Point,
		minAltitude,
		maxAltitude,
		radius unit.Length,
		bearing bearings.Bearing,
		arc unit.Angle,
		coalition coalitions.Coalition,
		category brevity.ContactCategory,
	) brevity.Group
	// SetFadedCallback sets the callback function to be called when a trackfile fades.
	SetFadedCallback(FadedCallback)
}

var _ Radar = &scope{}

type scope struct {
	updates       <-chan sim.Updated
	fades         <-chan sim.Faded
	missionTime   time.Time
	bullseye      orb.Point
	contacts      contactDatabase
	fadedCallback FadedCallback
}

func New(coalition coalitions.Coalition, updates <-chan sim.Updated, fades <-chan sim.Faded) Radar {
	return &scope{
		updates:       updates,
		fades:         fades,
		missionTime:   conf.InitialTime,
		contacts:      newContactDatabase(),
		fadedCallback: func(brevity.Group, coalitions.Coalition) {},
	}
}

func (s *scope) SetMissionTime(t time.Time) {
	s.missionTime = t
}

func (s *scope) SetBullseye(bullseye orb.Point) {
	s.bullseye = bullseye
}

func (s *scope) Bullseye() orb.Point {
	return s.bullseye
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
			s.handleFade(ctx, fade)
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
		Str("name", update.Labels.Name).
		Str("aircraft", update.Labels.ACMIName).
		Int("unitID", int(update.Labels.UnitID)).
		Logger()

	tf, ok := s.contacts.getByUnitID(update.Labels.UnitID)
	if ok {
		tf.Update(update.Frame)
		logger.Trace().Msg("updated existing trackfile")
	} else {
		tf = trackfiles.NewTrackfile(update.Labels)
		s.contacts.set(tf)
		logger.Info().Msg("created new trackfile")
	}
}

// handleFade removes any trackfiles for the faded unit.
func (s *scope) handleFade(ctx context.Context, fade sim.Faded) {
	trackfile, ok := s.contacts.getByUnitID(fade.UnitID)
	if !ok {
		log.Trace().Uint32("unitID", fade.UnitID).Msg("faded trackfile not found - probably not an aircraft")
		return
	}
	logger := log.With().
		Int("unitID", int(fade.UnitID)).
		Str("name", trackfile.Contact.Name).
		Str("aircraft", trackfile.Contact.ACMIName).
		Logger()

	if !ok {
		logger.Warn().Msg("faded trackfile not found")
		return
	}
	grp := newGroupUsingBullseye(s.bullseye)
	grp.contacts = append(grp.contacts, trackfile)
	s.contacts.delete(fade.UnitID)
	logger.Info().Msg("removed faded trackfile")
	if s.fadedCallback != nil {
		go func() {
			delay := time.Duration(rand.IntN(30)+30) * time.Second
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				log.Info().Msg("stopping FADED goroutine due to context cancellation")
				return
			case <-timer.C:
				if _, ok := s.contacts.getByUnitID(fade.UnitID); !ok {
					log.Info().Uint32("unitID", fade.UnitID).Str("group", grp.String()).Msg("calling faded callback for group")
					s.fadedCallback(grp, trackfile.Contact.Coalition)
				}
			}

		}()
	}
}

// handleGarbageCollection removes trackfiles that have not been updated in a long time.
func (s *scope) handleGarbageCollection() {
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		logger := log.With().
			Int("unitID", int(trackfile.Contact.UnitID)).
			Str("name", trackfile.Contact.Name).
			Str("aircraft", trackfile.Contact.ACMIName).
			Logger()

		lastSeen, ok := s.contacts.lastUpdated(trackfile.Contact.UnitID)
		if !ok {
			logger.Warn().Msg("last updated time is missing")
			continue
		}
		if lastSeen.Before(time.Now().Add(-5 * time.Minute)) {
			s.contacts.delete(trackfile.Contact.UnitID)
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
func isValidTrack(tf *trackfiles.Trackfile) bool {
	point := tf.LastKnown().Point
	isValidLongitude := point.Lon() != 0
	isValidLatitude := point.Lat() != 0
	isValidPosition := isValidLongitude && isValidLatitude
	isAboveSpeedFilter := tf.Speed() > 50*unit.Knot
	isAboveAltitudeFilter := tf.LastKnown().Altitude > 10*unit.Meter
	isValid := isValidPosition && isAboveSpeedFilter && isAboveAltitudeFilter
	log.Trace().
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
func (s *scope) isMatch(tf *trackfiles.Trackfile, coalition coalitions.Coalition, filter brevity.ContactCategory) bool {
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

func (s *scope) Declination(p orb.Point) unit.Angle {
	declination, err := bearings.Declination(p, s.missionTime)
	if err != nil {
		log.Error().Err(err).Msg("failed to get declination")
	}
	return declination
}
