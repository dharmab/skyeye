package radar

import (
	"context"
	"math"
	"slices"
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
)

type Radar interface {
	Run(context.Context)
	FindCallsign(string) *trackfile.Trackfile
	FindUnit(uint32) *trackfile.Trackfile
	GetBullseye() *orb.Point
	FindNearestGroup(orb.Point, common.Coalition, brevity.ContactCategory) brevity.Group
}

var _ Radar = &scope{}

type scope struct {
	simUpdates   <-chan dcs.Updated
	simFades     <-chan dcs.Faded
	contacts     map[string]*trackfile.Trackfile
	bullseye     *orb.Point
	aircraftData map[string]encyclopedia.Aircraft
}

func New() Radar {
	e := encyclopedia.New()

	return &scope{
		contacts:     make(map[string]*trackfile.Trackfile),
		aircraftData: e.Aircraft(),
	}
}

func (s *scope) Run(ctx context.Context) {
	for {
		select {
		case update := <-s.simUpdates:
			s.handleUpdate(update)
		case fade := <-s.simFades:
			s.handleFade(fade)
		case <-ctx.Done():
			return
		}
	}
}

func (s *scope) handleUpdate(update dcs.Updated) {
	tf, ok := s.contacts[update.Aircraft.Name]
	/// what if duplicate name tho
	if !ok {
		tf = trackfile.NewTrackfile(trackfile.Aircraft{
			UnitID:     update.Aircraft.UnitID,
			Name:       update.Aircraft.Name,
			Coalition:  update.Aircraft.Coalition,
			EditorType: update.Aircraft.EditorType,
		})
		s.contacts[update.Aircraft.Name] = tf
	} else {
		tf.Update(trackfile.Frame{
			Timestamp: time.Now(),
			Point:     update.Frame.Point,
			Altitude:  update.Frame.Altitude,
			Heading:   update.Frame.Heading,
			Speed:     update.Frame.Speed,
		})
	}
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

func (s *scope) FindNearestGroup(location orb.Point, coalition common.Coalition, filter brevity.ContactCategory) brevity.Group {
	var nearestTrackfile *trackfile.Trackfile
	nearestDistance := unit.Length(math.MaxFloat64)
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition {
			data, ok := s.aircraftData[tf.Contact.EditorType]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category == filter
			if matchesFilter {
				isNearer := planar.Distance(tf.Track.Front().Point, location) < nearestDistance.Meters()
				if nearestTrackfile == nil || isNearer {
					nearestTrackfile = tf
				}
			}
		}
	}

	return s.findGroupForAircraft(nearestTrackfile)
}

func (s *scope) findGroupForAircraft(trackfile *trackfile.Trackfile) brevity.Group {
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
func (s *scope) addNearbyAircraftToGroup(this *trackfile.Trackfile, group group) {
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
