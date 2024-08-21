package radar

import (
	"math"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
	"github.com/rs/zerolog/log"
)

// FindNearestTrackfile implements [Radar.FindNearestTrackfile]
func (s *scope) FindNearestTrackfile(
	origin orb.Point,
	minAltitude unit.Length,
	maxAltitude unit.Length,
	radius unit.Length,
	coalition coalitions.Coalition,
	filter brevity.ContactCategory,
) *trackfiles.Trackfile {
	var nearestTrackfile *trackfiles.Trackfile
	nearestDistance := radius
	for trackfile := range s.contacts.values() {
		isMatch := s.isMatch(trackfile, coalition, filter)
		altitude := trackfile.LastKnown().Altitude
		isWithinAltitude := minAltitude <= altitude && altitude <= maxAltitude
		if isMatch && isWithinAltitude {
			distance := spatial.Distance(origin, trackfile.LastKnown().Point)
			isNearer := distance < nearestDistance
			if isNearer {
				nearestTrackfile = trackfile
				nearestDistance = distance
			}
		}
	}
	if nearestTrackfile != nil {
		log.Debug().
			Any("origin", origin).
			Str("aircraft", nearestTrackfile.Contact.ACMIName).
			Int("unitID", int(nearestTrackfile.Contact.UnitID)).
			Int("altitude", int(nearestTrackfile.LastKnown().Altitude.Feet())).
			Msg("found nearest contact")
	} else {
		log.Debug().Msg("no contacts found within search volume")
	}
	return nearestTrackfile
}

// FindNearestGroupWithBRAA implements [Radar.FindNearestGroupWithBRAA]
func (s *scope) FindNearestGroupWithBRAA(
	origin orb.Point,
	minAltitude unit.Length,
	maxAltitude unit.Length,
	radius unit.Length,
	coalition coalitions.Coalition,
	filter brevity.ContactCategory,
) brevity.Group {
	trackfile := s.FindNearestTrackfile(origin, minAltitude, maxAltitude, radius, coalition, filter)
	if trackfile == nil {
		return nil
	}

	grp := s.findGroupForAircraft(trackfile)
	if grp == nil {
		return nil
	}

	declination := s.Declination(origin)
	bearing := spatial.TrueBearing(origin, grp.point()).Magnetic(declination)
	_range := spatial.Distance(origin, grp.point())
	aspect := brevity.AspectFromAngle(bearing, trackfile.Course())
	grp.braa = brevity.NewBRAA(
		bearing,
		_range,
		grp.altitudes(),
		aspect,
	)
	grp.bullseye = nil
	grp.aspect = &aspect
	grp.isThreat = _range < brevity.MandatoryThreatDistance

	return grp
}

// FindNearestGroupWithBullseye implements [Radar.FindNearestGroupWithBullseye]
func (s *scope) FindNearestGroupWithBullseye(origin orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, minAltitude, maxAltitude, radius, coalition, filter)
	grp := s.findGroupForAircraft(nearestTrackfile)
	declination := s.Declination(origin)
	bearing := spatial.TrueBearing(origin, grp.point()).Magnetic(declination)
	aspect := brevity.AspectFromAngle(bearing, grp.course())

	grp.aspect = &aspect
	_range := spatial.Distance(origin, grp.point())
	grp.isThreat = _range < brevity.MandatoryThreatDistance
	log.Debug().Any("origin", origin).Stringer("group", grp).Msg("determined nearest group")
	return grp
}

// FindNearestGroupInSector implements [Radar.FindNearestGroupInSector]
func (s *scope) FindNearestGroupInSector(origin orb.Point, minAltitude, maxAltitude, length unit.Length, bearing bearings.Bearing, arc unit.Angle, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	logger := log.With().Any("origin", origin).Stringer("bearing", bearing).Float64("arc", arc.Degrees()).Logger()

	declination := s.Declination(origin)
	bearing = bearing.Magnetic(declination)

	ring := orb.Ring{origin}
	for a := arc / 2; a > -arc/2; a -= arc / 10 {
		ring = append(
			ring,
			geo.PointAtBearingAndDistance(
				origin,
				(bearing.Value()+a).Degrees(),
				length.Meters(),
			),
		)
	}
	ring = append(ring, origin)
	sector := orb.Polygon{ring}

	logger.Debug().Any("sector", sector).Msg("searching sector")
	nearestDistance := unit.Length(math.MaxFloat64)
	var nearestContact *trackfiles.Trackfile
	for trackfile := range s.contacts.values() {
		logger := logger.With().Int("unitID", int(trackfile.Contact.UnitID)).Logger()
		isMatch := s.isMatch(trackfile, coalition, filter)
		isWithinAltitude := minAltitude <= trackfile.LastKnown().Altitude && trackfile.LastKnown().Altitude <= maxAltitude
		if isMatch && isWithinAltitude {
			contactLocation := trackfile.LastKnown().Point
			distanceToContact := spatial.Distance(origin, contactLocation)
			inSector := planar.PolygonContains(sector, contactLocation)
			logger.Debug().Float64("distanceNM", distanceToContact.NauticalMiles()).Bool("inSector", inSector).Msg("checking distance and location")
			if distanceToContact < nearestDistance && distanceToContact > conf.DefaultMarginRadius && inSector {
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
	grp := s.findGroupForAircraft(nearestContact)
	if grp == nil {
		return nil
	}
	preciseBearing := spatial.TrueBearing(origin, nearestContact.LastKnown().Point).Magnetic(declination)
	aspect := brevity.AspectFromAngle(preciseBearing, nearestContact.Course())
	log.Debug().Str("aspect", string(aspect)).Msg("determined aspect")
	_range := spatial.Distance(origin, nearestContact.LastKnown().Point)
	grp.aspect = &aspect
	grp.braa = brevity.NewBRAA(
		preciseBearing,
		_range,
		grp.altitudes(),
		grp.Aspect(),
	)
	logger.Debug().Stringer("group", grp).Msg("determined nearest group")
	grp.bullseye = nil
	return grp
}
