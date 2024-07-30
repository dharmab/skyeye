package radar

import (
	"math"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/planar"
	"github.com/rs/zerolog/log"
)

func (s *scope) FindNearestTrackfile(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) *trackfile.Trackfile {
	var nearestTrackfile *trackfile.Trackfile
	nearestDistance := unit.Length(300) * unit.NauticalMile
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		if tf.Contact.Coalition == coalition && isValidTrack(tf) {
			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category() == filter || filter == brevity.Aircraft
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

func (s *scope) FindNearestGroupWithBullseye(origin orb.Point, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := s.FindNearestTrackfile(origin, coalition, filter)
	group := s.findGroupForAircraft(nearestTrackfile)
	groupLocation := nearestTrackfile.LastKnown().Point
	aspect := brevity.AspectFromAngle(unit.Angle(geo.Bearing(origin, groupLocation))*unit.Degree, nearestTrackfile.Course())
	group.aspect = &aspect
	rang := unit.Length(geo.Distance(origin, groupLocation)) * unit.Meter
	group.isThreat = rang < brevity.MandatoryThreatDistance
	log.Debug().Interface("origin", origin).Interface("group", group).Msg("determined nearest group")
	return group
}

func (s *scope) FindNearestGroupInCone(origin orb.Point, bearing unit.Angle, arc unit.Angle, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	maxDistance := conf.DefaultPictureRadius
	cone := orb.Polygon{
		orb.Ring{
			origin,
			geo.PointAtBearingAndDistance(origin, (bearing + (arc / 2)).Degrees(), maxDistance.Meters()),
			geo.PointAtBearingAndDistance(origin, (bearing - (arc / 2)).Degrees(), maxDistance.Meters()),
			origin,
		},
	}
	logger := log.With().Interface("origin", origin).Float64("bearing", bearing.Degrees()).Float64("arc", arc.Degrees()).Logger()
	logger.Debug().Any("cone", cone).Msg("searching cone")

	nearestDistance := unit.Length(math.MaxFloat64)
	var nearestContact *trackfile.Trackfile
	s.lock.Lock()
	defer s.lock.Unlock()
	for unitID, tf := range s.contacts {
		logger := logger.With().Int("unitID", int(unitID)).Logger()
		if tf.Contact.Coalition == coalition {
			logger.Debug().Msg("checking contact")
			if !isValidTrack(tf) {
				logger.Debug().Msg("contact is below speed threshold")
				continue
			}

			data, ok := s.aircraftData[tf.Contact.ACMIName]
			// If the aircraft is not in the encyclopedia, assume it matches
			matchesFilter := !ok || data.Category() == filter || filter == brevity.Aircraft
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

	logger = log.With().Int("unitID", int(nearestContact.Contact.UnitID)).Logger()
	logger.Debug().Msg("found nearest contact")
	group := s.findGroupForAircraft(nearestContact)
	if group == nil {
		return nil
	}
	exactBearing := unit.Angle(geo.Bearing(origin, nearestContact.LastKnown().Point)) * unit.Degree
	aspect := brevity.AspectFromAngle(bearing, nearestContact.Course())
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
