package radar

import (
	"context"
	"math"
	"slices"
	"strings"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
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
	// GetBullseye returns the bullseye for the given coalition.
	GetBullseye(types.Coalition) dcs.Bullseye
	// FindNearbyGroups returns all groups within 3 nautical miles of the given location, filtered by the given contact category.
	// Location data is unset, since it is within radar margins of the given location.
	FindNearbyGroups(orb.Point, types.Coalition, brevity.ContactCategory) []brevity.Group
	// FindNearestGroupWithBRAA returns the nearest group to the given location, with BRAA location embedded in the Group.
	// The given point is the location to search from.
	// The given coalition is the coalition to search for.
	// The given filter is the contact category to filter by.
	// Returns the nearest group to the given location which matches the given coalition and filter, with BRAA relative to the given location. Returns nil if no group was found.
	FindNearestGroupWithBRAA(orb.Point, types.Coalition, brevity.ContactCategory) brevity.Group
	// FindNearestGroupWithBullseye returns the nearest group to the given location, with Bullseye location embedded in the Group.
	// The given point is the location to search from.
	// The given coalition is the coalition to search for.
	// The given filter is the contact category to filter by.
	// Returns the nearest group to the given location which matches the given coalition and filter, with Bullseye location. Returns nil if no group was found.
	FindNearestGroupWithBullseye(orb.Point, types.Coalition, brevity.ContactCategory) brevity.Group
	// FindNearestGroupInCone returns the nearest group to the given location along the given bearing, ± the given angle, with BRAA relative to the given location. Returns nil if no group was found.
	FindNearestGroupInCone(orb.Point, unit.Angle, unit.Angle, types.Coalition, brevity.ContactCategory) brevity.Group
}

var _ Radar = &scope{}

type scope struct {
	simUpdates   <-chan dcs.Updated
	simFades     <-chan dcs.Faded
	simBullseyes <-chan dcs.Bullseye
	contacts     map[string]*trackfile.Trackfile
	bullseyes    map[types.Coalition]dcs.Bullseye
	aircraftData map[string]encyclopedia.Aircraft
}

func New(coalition types.Coalition, bullseyes <-chan dcs.Bullseye, updates <-chan dcs.Updated, fades <-chan dcs.Faded) Radar {
	e := encyclopedia.New()

	return &scope{
		simUpdates:   updates,
		simFades:     fades,
		simBullseyes: bullseyes,
		bullseyes:    make(map[types.Coalition]dcs.Bullseye),
		contacts:     make(map[string]*trackfile.Trackfile),
		aircraftData: e.Aircraft(),
	}
}

func (s *scope) Run(ctx context.Context) {
	for {
		select {
		case bullseye := <-s.simBullseyes:
			s.bullseyes[bullseye.Coalition] = bullseye
		case update := <-s.simUpdates:
			s.handleUpdate(update)
		case fade := <-s.simFades:
			s.handleFade(fade)
		case <-ctx.Done():
			return
		}
	}
}

func (s *scope) RunOnce() {
	for {
		select {
		case bullseye := <-s.simBullseyes:
			s.bullseyes[bullseye.Coalition] = bullseye
		case update := <-s.simUpdates:
			s.handleUpdate(update)
		case fade := <-s.simFades:
			s.handleFade(fade)
		default:
			return
		}
	}
}

func (s *scope) handleUpdate(update dcs.Updated) {
	callsign, _, _ := strings.Cut(update.Aircraft.Name, " | ")
	// replace digits and spaces with digit followed by a single space
	callsign, ok := parser.ParseCallsign(callsign)

	if !ok {
		callsign = update.Aircraft.Name
	}

	_, ok = s.contacts[callsign]
	/// what if duplicate callsign tho
	if !ok {
		s.contacts[callsign] = trackfile.NewTrackfile(update.Aircraft)
	}
	s.contacts[callsign].Update(update.Frame)

}

func (s *scope) handleFade(fade dcs.Faded) {
	// TODO mark faded? Move from contacts to fadedContacts?
}

func (s *scope) FindCallsign(callsign string) *trackfile.Trackfile {
	tf, ok := s.contacts[callsign]
	if !ok {
		return nil
	}
	return tf
}

func (s *scope) FindUnit(unitId uint32) *trackfile.Trackfile {
	for _, tf := range s.contacts {
		if tf.Contact.UnitID == unitId {
			return tf
		}
	}
	return nil
}

func (s *scope) GetBullseye(coalition types.Coalition) dcs.Bullseye {
	return s.bullseyes[coalition]
}

func (s *scope) FindNearestGroupWithBRAA(location orb.Point, coalition types.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(location, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	bearing := unit.Angle(geo.Bearing(location, groupLocation)) * unit.Degree
	_range := unit.Length(planar.Distance(location, groupLocation)) * unit.Meter
	altitude := nearestTrackfile.LastKnown().Altitude
	aspect := brevity.AspectFromAngle(bearing, nearestTrackfile.LastKnown().Heading)
	group.braa = brevity.NewBRAA(
		bearing,
		_range,
		altitude,
		aspect,
	)
	group.aspect = &aspect
	group.isThreat = _range < brevity.MandatoryThreatDistance
	return group
}

func (s *scope) FindNearestGroupWithBullseye(location orb.Point, coalition types.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(location, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	aspect := brevity.AspectFromAngle(unit.Angle(geo.Bearing(location, groupLocation))*unit.Degree, nearestTrackfile.LastKnown().Heading)
	group.aspect = &aspect
	rang := unit.Length(planar.Distance(location, groupLocation)) * unit.Meter
	group.isThreat = rang < brevity.MandatoryThreatDistance
	return group
}

func (s *scope) FindNearestTrackfile(location orb.Point, coalition types.Coalition, filter brevity.ContactCategory) *trackfile.Trackfile {
	var nearestTrackfile *trackfile.Trackfile
	nearestDistance := unit.Length(math.MaxFloat64)
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			data, ok := s.aircraftData[tf.Contact.EditorType]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Everything
			if matchesFilter {
				isNearer := planar.Distance(tf.LastKnown().Point, location) < nearestDistance.Meters()
				if nearestTrackfile == nil || isNearer {
					nearestTrackfile = tf
				}
			}
		}
	}
	return nearestTrackfile
}

func (s *scope) FindNearbyGroups(location orb.Point, coalition types.Coalition, filter brevity.ContactCategory) []brevity.Group {
	groups := make([]brevity.Group, 0)
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			data, ok := s.aircraftData[tf.Contact.EditorType]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Everything
			if matchesFilter {
				if planar.Distance(tf.LastKnown().Point, location) < conf.DefaultMarginRadius.Meters() {
					group := s.findGroupForAircraft(tf)
					groups = append(groups, group)
				}
			}
		}
	}
	return groups
}

func (s *scope) FindNearestGroupInCone(location orb.Point, bearing unit.Angle, width unit.Angle, coalition types.Coalition, filter brevity.ContactCategory) brevity.Group {
	maxDistance := 120 * unit.NauticalMile
	cone := orb.Polygon{
		orb.Ring{
			location,
			geo.PointAtBearingAndDistance(location, (bearing - width/2).Degrees(), maxDistance.Meters()),
			geo.PointAtBearingAndDistance(location, (bearing + width/2).Degrees(), maxDistance.Meters()),
			location,
		},
	}

	distanceFromLocation := unit.Length(math.MaxFloat64)
	var nearestContact *trackfile.Trackfile
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			data, ok := s.aircraftData[tf.Contact.EditorType]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter || filter == brevity.Everything
			if matchesFilter {
				if planar.Distance(tf.LastKnown().Point, location) < distanceFromLocation.Meters() {
					if planar.PolygonContains(cone, tf.LastKnown().Point) {
						nearestContact = tf
					}
				}
			}
		}
	}
	if nearestContact == nil {
		return nil
	} else {
		return s.findGroupForAircraft(nearestContact)
	}
}

func (s *scope) findGroupForAircraft(trackfile *trackfile.Trackfile) *group {
	if trackfile == nil {
		return nil
	}
	group := newGroupUsingBullseye(s.bullseyes[trackfile.Contact.Coalition].Point)
	group.contacts = append(group.contacts, trackfile)
	s.addNearbyAircraftToGroup(trackfile, group)
	return group
}

// addNearbyAircraftToGroup recursively adds all nearby aircraft which:
//
// - are of the same coalition
//
// - are of the same platform
//
// - are within 1 nautical mile in 2D distance of each other
//
// - are within 1000 feet in altitude of each other
func (s *scope) addNearbyAircraftToGroup(this *trackfile.Trackfile, group *group) {
	spreadInterval := unit.Length(1) * unit.NauticalMile
	stackInterval := unit.Length(1000) * unit.Foot
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

		// Compare attributes that are shared within a group
		isSameCoalition := other.Contact.Coalition == this.Contact.Coalition
		isSamePlatform := s.aircraftData[other.Contact.EditorType].PlatformDesignation == s.aircraftData[this.Contact.EditorType].PlatformDesignation

		isWithinSpread := planar.Distance(other.LastKnown().Point, this.LastKnown().Point) < spreadInterval.Meters()
		isWithinStack := math.Abs(other.LastKnown().Altitude.Feet()-this.LastKnown().Altitude.Feet()) < stackInterval.Feet()
		if isSameCoalition && isSamePlatform && isWithinSpread && isWithinStack {
			if planar.Distance(other.LastKnown().Point, this.LastKnown().Point) < spreadInterval.Meters() {
				group.contacts = append(group.contacts, other)
				s.addNearbyAircraftToGroup(other, group)
			}
		}
	}
}
