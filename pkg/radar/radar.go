package radar

import (
	"context"
	"math"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
	"github.com/rs/zerolog/log"
)

type Radar interface {
	// Run consumes updates from the simulation channels until the context is cancelled.
	Run(context.Context)
	// RunOnce consumes all updates from the simulation channels, then exits. It is intended for use in tests, in combination with buffered channels preloaded with test data.
	RunOnce()
	// FindCallsign returns the trackfile for the given callsign, or nil if no trackfile was found.
	FindCallsign(string) *trackfile.Trackfile
	// FindUnit returns the trackfile for the given unit ID, or nil if no trackfile was found.
	FindUnit(uint32) *trackfile.Trackfile
	// GetBullseye returns the bullseye for the configured coalition.
	GetBullseye() orb.Point
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
	for {
		select {
		case update := <-s.updates:
			s.handleUpdate(update)
		case fade := <-s.fades:
			s.handleFade(fade)
		case bullseye := <-s.bullseyes:
			s.bullseye = bullseye
		case <-ctx.Done():
			return
		}
	}
}

func (s *scope) RunOnce() {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case bullseye := <-s.bullseyes:
			s.bullseye = bullseye
		case update := <-s.updates:
			s.handleUpdate(update)
		case fade := <-s.fades:
			s.handleFade(fade)
		case <-ticker.C:
			s.handleGarbageCollection()
		default:
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

func (s *scope) FindCallsign(callsign string) *trackfile.Trackfile {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Debug().Str("callsign", callsign).Any("contacts", s.contacts).Msg("searching scope for trackfile matching callsign")
	unitID, ok := s.callsignIdx[callsign]
	if !ok {
		return nil
	}
	tf, ok := s.contacts[unitID]
	if !ok {
		return nil
	}
	return tf
}

func (s *scope) FindUnit(unitId uint32) *trackfile.Trackfile {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.UnitID == unitId {
			return tf
		}
	}
	return nil
}

func (s *scope) GetBullseye() orb.Point {
	return s.bullseye
}

func (s *scope) FindNearestGroupWithBRAA(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	if group == nil {
		return nil
	}
	groupLocation := nearestTrackfile.LastKnown().Point
	bearing := unit.Angle(geo.Bearing(origin, groupLocation)) * unit.Degree
	_range := unit.Length(geo.Distance(origin, groupLocation)) * unit.Meter
	altitude := nearestTrackfile.LastKnown().Altitude
	aspect := brevity.AspectFromAngle(bearing, nearestTrackfile.LastKnown().Heading)
	group.track = nearestTrackfile.Direction()
	group.braa = brevity.NewBRAA(
		bearing,
		_range,
		altitude,
		aspect,
	)
	group.bullseye = nil
	group.aspect = &aspect
	group.isThreat = _range < brevity.MandatoryThreatDistance

	return group
}

func (s *scope) FindNearestGroupWithBullseye(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	aspect := brevity.AspectFromAngle(unit.Angle(geo.Bearing(origin, groupLocation))*unit.Degree, nearestTrackfile.LastKnown().Heading)
	group.aspect = &aspect
	group.track = nearestTrackfile.Direction()
	rang := unit.Length(geo.Distance(origin, groupLocation)) * unit.Meter
	group.isThreat = rang < brevity.MandatoryThreatDistance
	log.Debug().Interface("origin", origin).Interface("group", group).Msg("determined nearest group")
	return group
}

func (s *scope) FindNearestTrackfile(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) *trackfile.Trackfile {
	var nearestTrackfile *trackfile.Trackfile
	nearestDistance := unit.Length(300) * unit.NauticalMile
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition && isValidTrack(tf) {
			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Aircraft
			if matchesFilter {
				hasTrack := tf.Track.Len() > 0
				if hasTrack {
					distance := unit.Length(math.Abs(geo.Distance(origin, tf.LastKnown().Point)))
					isNearer := distance < nearestDistance
					if nearestTrackfile == nil || isNearer {
						log.Debug().
							Interface("origin", origin).
							Int("distance", int(distance.NauticalMiles())).
							Str("aircraft", tf.Contact.ACMIName).
							Int("unitID", int(tf.Contact.UnitID)).
							Str("name", tf.Contact.Name).
							Msg("new candidate for nearest trackfile")
						nearestTrackfile = tf
						nearestDistance = distance
					}
				}
			}
		}
	}
	if nearestTrackfile != nil {
		log.Debug().
			Interface("origin", origin).
			Str("aircraft", nearestTrackfile.Contact.ACMIName).
			Int("unitID", int(nearestTrackfile.Contact.UnitID)).
			Int("altitude", int(nearestTrackfile.LastKnown().Altitude.Feet())).
			Msg("found nearest contact")
	}
	return nearestTrackfile
}

func (s *scope) FindNearbyGroups(location orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) []brevity.Group {
	groups := make([]brevity.Group, 0)
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			if !isValidTrack(tf) {
				continue
			}
			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Aircraft
			if matchesFilter {
				if geo.Distance(tf.LastKnown().Point, location) < conf.DefaultMarginRadius.Meters() {
					group := s.findGroupForAircraft(tf)
					groups = append(groups, group)
				}
			}
		}
	}
	return groups
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

func (s *scope) FindNearestGroupInCone(origin orb.Point, bearing unit.Angle, arc unit.Angle, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	maxDistance := 150 * unit.NauticalMile
	cone := orb.Polygon{
		orb.Ring{
			origin,
			geo.PointAtBearingAndDistance(origin, 180+(bearing-(arc/2)).Degrees(), maxDistance.Meters()),
			geo.PointAtBearingAndDistance(origin, 180+(bearing+(arc/2)).Degrees(), maxDistance.Meters()),
			origin,
		},
	}

	nearestDistance := unit.Length(math.MaxFloat64)
	var nearestContact *trackfile.Trackfile
	s.lock.Lock()
	defer s.lock.Unlock()
	for unitID, tf := range s.contacts {
		logger := log.With().Int("unitID", int(unitID)).Logger()
		if tf.Contact.Coalition == coalition {
			logger.Debug().Msg("checking contact")
			if !isValidTrack(tf) {
				logger.Debug().Msg("contact is below speed threshold")
				continue
			}

			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Aircraft
			if matchesFilter {
				logger.Debug().Msg("contact matches filter")
				contactLocation := tf.LastKnown().Point
				distanceToContact := unit.Length(geo.Distance(origin, contactLocation)) * unit.Meter
				isWithinCone := planar.PolygonContains(cone, contactLocation)
				logger.Debug().Float64("distanceNM", distanceToContact.NauticalMiles()).Bool("isWithinCone", isWithinCone).Msg("checking distance and location")
				if distanceToContact < nearestDistance && distanceToContact > conf.DefaultMarginRadius && isWithinCone {
					nearestContact = tf
				}
			}
		}
	}
	if nearestContact == nil {
		log.Debug().Msg("no contacts found in cone")
		return nil
	}

	logger := log.With().Int("unitID", int(nearestContact.Contact.UnitID)).Logger()
	logger.Debug().Msg("found nearest contact")
	group := s.findGroupForAircraft(nearestContact)
	if group == nil {
		return nil
	}
	exactBearing := unit.Angle(geo.Bearing(origin, nearestContact.LastKnown().Point)) * unit.Degree
	aspect := brevity.AspectFromAngle(bearing, nearestContact.LastKnown().Heading)
	log.Debug().Str("aspect", string(aspect)).Msg("determined aspect")
	_range := unit.Length(geo.Distance(origin, nearestContact.LastKnown().Point)) * unit.Meter
	group.aspect = &aspect
	group.track = nearestContact.Direction()
	group.braa = brevity.NewBRAA(
		exactBearing,
		_range,
		group.Altitude(),
		group.Aspect(),
	)
	logger.Debug().Interface("group", group).Msg("determined nearest group")
	group.bullseye = nil
	return group

}

func (s *scope) findGroupForAircraft(tf *trackfile.Trackfile) *group {
	if tf == nil {
		return nil
	}
	group := newGroupUsingBullseye(s.bullseye)
	group.contacts = append(group.contacts, tf)
	s.addNearbyAircraftToGroup(tf, group)
	platforms := make(map[string]any)
	for _, tf := range group.contacts {
		var name string
		data, ok := s.aircraftData[tf.Contact.ACMIName]
		if ok {
			for _, reportingName := range []string{data.NATOReportingName, data.Nickname, data.OfficialName, data.PlatformDesignation} {
				if reportingName != "" {
					name = reportingName
					break
				}
			}
		}
		platforms[name] = nil
	}
	for platform := range platforms {
		group.platforms = append(group.platforms, platform)
	}

	return group
}

// addNearbyAircraftToGroup recursively adds all nearby aircraft which:
//
// - are of the same coalition
//
// - are within 3 nautical miles in 2D distance of each other
//
// - are within 3000 feet in altitude of each other
//
// These are tripled from the ATP numbers beacause the DCS AI isn't amazing at holding formation.
// We allow mixed platform groups because these are fairly common in DCS.
func (s *scope) addNearbyAircraftToGroup(this *trackfile.Trackfile, group *group) {
	spreadInterval := unit.Length(3) * unit.NauticalMile
	stackInterval := unit.Length(3000) * unit.Foot
	for _, other := range s.contacts {
		// Skip if this one is already in the group
		if slices.ContainsFunc(group.contacts, func(t *trackfile.Trackfile) bool {
			if t == nil {
				return false
			}
			return t.Contact.UnitID == other.Contact.UnitID
		}) {
			continue
		}

		if !isValidTrack(other) {
			continue
		}

		isSameCoalition := other.Contact.Coalition == this.Contact.Coalition
		isWithinSpread := geo.Distance(other.LastKnown().Point, this.LastKnown().Point) < spreadInterval.Meters()
		isWithinStack := math.Abs(other.LastKnown().Altitude.Feet()-this.LastKnown().Altitude.Feet()) < stackInterval.Feet()
		log.Debug().
			Any("initialContact", this.Contact).
			Any("contact", other.Contact).
			Int("unitID", int(other.Contact.UnitID)).
			Bool("isSameCoalition", isSameCoalition).
			Bool("isWithinSpread", isWithinSpread).
			Bool("isWithinStack", isWithinStack).
			Msg("checking if contact is within group")
		if isSameCoalition && isWithinSpread && isWithinStack {
			group.contacts = append(group.contacts, other)
			s.addNearbyAircraftToGroup(other, group)
		}
	}
}
