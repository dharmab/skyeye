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

// FindNearestTrackfile returns the nearest trackfile to the given origin (up to the given radius), within the
// given altitude block, filtered by the given coalition and contact category. Returns nil if no trackfile was found.
func (r *Radar) FindNearestTrackfile(
	origin orb.Point,
	minAltitude unit.Length,
	maxAltitude unit.Length,
	radius unit.Length,
	coalition coalitions.Coalition,
	filter brevity.ContactCategory,
) *trackfiles.Trackfile {
	var nearestTrackfile *trackfiles.Trackfile
	nearestDistance := radius
	for trackfile := range r.contacts.values() {
		isMatch := isMatch(trackfile, coalition, filter)
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
			Uint64("id", nearestTrackfile.Contact.ID).
			Int("altitude", int(nearestTrackfile.LastKnown().Altitude.Feet())).
			Msg("found nearest contact")
	} else {
		log.Debug().Msg("no contacts found within search volume")
	}
	return nearestTrackfile
}

// FindNearestGroupWithBRAA returns the nearest group to the given origin (up to the given radius), within the
// given altitude block, filtered by the given coalition and contact category. The group has BRAA set relative to
// the given origin. Returns nil if no group was found.
func (r *Radar) FindNearestGroupWithBRAA(
	origin orb.Point,
	minAltitude unit.Length,
	maxAltitude unit.Length,
	radius unit.Length,
	coalition coalitions.Coalition,
	filter brevity.ContactCategory,
) brevity.Group {
	trackfile := r.FindNearestTrackfile(origin, minAltitude, maxAltitude, radius, coalition, filter)
	if trackfile == nil || trackfile.IsLastKnownPointZero() {
		return nil
	}

	grp := r.findGroupForAircraft(trackfile)
	if grp == nil {
		return nil
	}

	declination := r.Declination(trackfile.LastKnown().Point)
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
	grp.isThreat = r.isGroupWithinThreatRadius(grp, _range)

	return grp
}

// FindNearestGroupWithBullseye returns the nearest group to the given origin (up to the given distance), within a 2D
// circular sector defined by the given origin ,radius, bearing and arc, within the given altitude block, filtered
// by the given coalition and contact category. The group has BRAA set relative to the given origin. Returns nil if
// no group was found.
func (r *Radar) FindNearestGroupWithBullseye(origin orb.Point, minAltitude, maxAltitude, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	nearestTrackfile := r.FindNearestTrackfile(origin, minAltitude, maxAltitude, radius, coalition, filter)
	grp := r.findGroupForAircraft(nearestTrackfile)
	declination := r.Declination(nearestTrackfile.LastKnown().Point)
	bearing := spatial.TrueBearing(origin, grp.point()).Magnetic(declination)
	aspect := brevity.AspectFromAngle(bearing, grp.course())

	grp.aspect = &aspect
	_range := spatial.Distance(origin, grp.point())
	grp.isThreat = r.isGroupWithinThreatRadius(grp, _range)
	log.Debug().Any("origin", origin).Stringer("group", grp).Msg("determined nearest group")
	return grp
}

// FindNearestGroupInSector returns the nearest group to the given origin within a 2D sector defined by the given
// origin, length, bearing and arc, within the given altitude block, filtered by the given coalition and contact category.
// The group has BRAA set relative to the given origin. Returns nil if no group was found.
func (r *Radar) FindNearestGroupInSector(origin orb.Point, minAltitude, maxAltitude, length unit.Length, bearing bearings.Bearing, arc unit.Angle, coalition coalitions.Coalition, filter brevity.ContactCategory) brevity.Group {
	logger := log.With().Any("origin", origin).Stringer("bearing", bearing).Float64("arc", arc.Degrees()).Logger()

	declination := r.Declination(origin)
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
	for trackfile := range r.contacts.values() {
		logger := logger.With().Uint64("id", trackfile.Contact.ID).Logger()
		isMatch := isMatch(trackfile, coalition, filter)
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

	logger = log.With().Uint64("id", nearestContact.Contact.ID).Logger()
	logger.Debug().Msg("found nearest contact")
	grp := r.findGroupForAircraft(nearestContact)
	if grp == nil {
		return nil
	}
	preciseBearing := spatial.TrueBearing(origin, nearestContact.LastKnown().Point).Magnetic(r.Declination(nearestContact.LastKnown().Point))
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
