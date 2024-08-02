package radar

import (
	"math"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
	"github.com/rs/zerolog/log"
)

// FindNearestTrackfile implements [Radar.FindNearestTrackfile]
func (s *scope) FindNearestTrackfile(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) *trackfiles.Trackfile {
	var nearestTrackfile *trackfiles.Trackfile
	nearestDistance := 300 * unit.NauticalMile
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		if s.isMatch(trackfile, coalition, filter) {
			distance := unit.Length(math.Abs(geo.Distance(origin, trackfile.LastKnown().Point)))
			isNearer := distance < nearestDistance
			if nearestTrackfile == nil || isNearer {
				log.Debug().
					Interface("origin", origin).
					Int("distance", int(distance.NauticalMiles())).
					Str("aircraft", trackfile.Contact.ACMIName).
					Int("unitID", int(trackfile.Contact.UnitID)).
					Str("name", trackfile.Contact.Name).
					Msg("new candidate for nearest trackfile")
				nearestTrackfile = trackfile
				nearestDistance = distance
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

// FindNearestGroupWithBRAA implements [Radar.FindNearestGroupWithBRAA]
func (s *scope) FindNearestGroupWithBRAA(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	if group == nil {
		return nil
	}
	groupLocation := nearestTrackfile.LastKnown().Point

	bearing := bearings.NewTrueBearing(
		unit.Angle(
			geo.Bearing(origin, groupLocation),
		) * unit.Degree,
	).Magnetic(s.bestAvailableDeclination(origin))
	_range := unit.Length(geo.Distance(origin, groupLocation)) * unit.Meter
	altitude := nearestTrackfile.LastKnown().Altitude
	aspect := brevity.AspectFromAngle(bearing, nearestTrackfile.Course())
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

// FindNearestGroupWithBullseye implements [Radar.FindNearestGroupWithBullseye]
func (s *scope) FindNearestGroupWithBullseye(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	aspect := brevity.AspectFromAngle(
		bearings.NewTrueBearing(
			unit.Angle(
				geo.Bearing(origin, groupLocation),
			)*unit.Degree,
		).Magnetic(s.bestAvailableDeclination(origin)), nearestTrackfile.Course(),
	)

	group.aspect = &aspect
	rang := unit.Length(geo.Distance(origin, groupLocation)) * unit.Meter
	group.isThreat = rang < brevity.MandatoryThreatDistance
	log.Debug().Interface("origin", origin).Interface("group", group).Msg("determined nearest group")
	return group
}

func (s *scope) FindNearestGroupInCone(origin orb.Point, bearing bearings.Bearing, arc unit.Angle, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	logger := log.With().Interface("origin", origin).Float64("bearing", bearing.Degrees()).Float64("arc", arc.Degrees()).Logger()
	maxDistance := conf.DefaultPictureRadius
	declination := s.bestAvailableDeclination(origin)
	vertex := func(a unit.Angle) orb.Point {
		return geo.PointAtBearingAndDistance(
			origin,
			(bearing.Magnetic(declination).Value() + a).Degrees(),
			maxDistance.Meters(),
		)
	}
	cone := orb.Polygon{
		orb.Ring{
			origin,
			vertex(arc / 2),
			vertex(-arc / 2),
			origin,
		},
	}
	logger.Debug().Any("cone", cone).Msg("searching cone")

	nearestDistance := unit.Length(math.MaxFloat64)
	var nearestContact *trackfiles.Trackfile
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		logger := logger.With().Int("unitID", int(trackfile.Contact.UnitID)).Logger()
		if s.isMatch(trackfile, coalition, filter) {
			contactLocation := trackfile.LastKnown().Point
			distanceToContact := unit.Length(geo.Distance(origin, contactLocation)) * unit.Meter
			isWithinCone := planar.PolygonContains(cone, contactLocation)
			logger.Debug().Float64("distanceNM", distanceToContact.NauticalMiles()).Bool("isWithinCone", isWithinCone).Msg("checking distance and location")
			if distanceToContact < nearestDistance && distanceToContact > conf.DefaultMarginRadius && isWithinCone {
				nearestContact = trackfile
			}
		}
	}
	if nearestContact == nil {
		log.Debug().Msg("no contacts found in cone")
		return nil
	}

	logger = log.With().Int("unitID", int(nearestContact.Contact.UnitID)).Logger()
	logger.Debug().Msg("found nearest contact")
	group := s.findGroupForAircraft(nearestContact)
	if group == nil {
		return nil
	}
	exactBearing := bearings.NewTrueBearing(
		unit.Angle(
			geo.Bearing(origin, nearestContact.LastKnown().Point),
		) * unit.Degree,
	).Magnetic(declination)
	aspect := brevity.AspectFromAngle(exactBearing, nearestContact.Course())
	log.Debug().Str("aspect", string(aspect)).Msg("determined aspect")
	_range := unit.Length(geo.Distance(origin, nearestContact.LastKnown().Point)) * unit.Meter
	group.aspect = &aspect
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
