package radar

import (
	"context"
	"math"
	"slices"
	"strings"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/parser"
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
	FindCallsign(string) *trackfile.Trackfile
	FindUnit(uint32) *trackfile.Trackfile
	GetBullseye() *orb.Point
	FindNearestGroupWithBRAA(orb.Point, common.Coalition, brevity.ContactCategory) brevity.Group
	FindNearestGroupWithBullseye(orb.Point, common.Coalition, brevity.ContactCategory) brevity.Group
}

var _ Radar = &scope{}

type scope struct {
	simUpdates   <-chan dcs.Updated
	simFades     <-chan dcs.Faded
	simBullseyes <-chan orb.Point
	contacts     map[string]*trackfile.Trackfile
	bullseye     *orb.Point
	aircraftData map[string]encyclopedia.Aircraft
}

func New(bullseyes <-chan orb.Point, updates <-chan dcs.Updated, fades <-chan dcs.Faded) Radar {
	e := encyclopedia.New()

	return &scope{
		simUpdates:   updates,
		simFades:     fades,
		simBullseyes: bullseyes,
		bullseye:     &orb.Point{0, 0},
		contacts:     make(map[string]*trackfile.Trackfile),
		aircraftData: e.Aircraft(),
	}
}

func (s *scope) Run(ctx context.Context) {
	for {
		select {
		case bullseye := <-s.simBullseyes:
			s.bullseye = &bullseye
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
			s.bullseye = &bullseye
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

func (s *scope) GetBullseye() *orb.Point {
	return s.bullseye
}

func (s *scope) FindNearestGroupWithBRAA(location orb.Point, coalition common.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(location, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	bearing := unit.Angle(geo.Bearing(location, groupLocation)) * unit.Degree
	rang := unit.Length(planar.Distance(location, groupLocation)) * unit.Meter
	altitude := nearestTrackfile.LastKnown().Altitude
	aspect := brevity.AspectFromAngle(bearing, nearestTrackfile.LastKnown().Heading)
	group.braa = brevity.NewBRAA(
		bearing,
		rang,
		altitude,
		aspect,
	)
	group.aspect = &aspect
	group.isThreat = rang < brevity.MandatoryThreatDistance
	return group
}

func (s *scope) FindNearestGroupWithBullseye(location orb.Point, coalition common.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(location, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	aspect := brevity.AspectFromAngle(unit.Angle(geo.Bearing(location, groupLocation))*unit.Degree, nearestTrackfile.LastKnown().Heading)
	group.aspect = &aspect
	rang := unit.Length(planar.Distance(location, groupLocation)) * unit.Meter
	group.isThreat = rang < brevity.MandatoryThreatDistance
	return group
}

func (s *scope) FindNearestTrackfile(location orb.Point, coalition common.Coalition, filter brevity.ContactCategory) *trackfile.Trackfile {
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

func (s *scope) findGroupForAircraft(trackfile *trackfile.Trackfile) *group {
	if trackfile == nil {
		return nil
	}
	group := newGroupUsingBullseye(*s.bullseye)
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
