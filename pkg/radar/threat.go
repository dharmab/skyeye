package radar

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

// Threats returns a map of hostile groups of the given coalition to the friendly object IDs that
// will hear the threat call. The receiver list is filtered to friendlies that are on the
// controller's SRS frequency, if an SRS client is available.
func (r *Radar) Threats(coalition coalitions.Coalition) map[brevity.Group][]uint64 {
	threats := make(map[*group][]uint64)
	hostileGroups := r.enumerateGroups(coalition)
	radius := 100 * unit.NauticalMile
	if r.mandatoryThreatRadius > radius {
		radius = r.mandatoryThreatRadius
	}
	for _, grp := range hostileGroups {
		friendlyGroups := r.findNearbyGroups(
			grp.point(),
			0,
			math.MaxFloat64,
			radius,
			coalition.Opposite(),
			brevity.Aircraft,
			[]uint64{},
		)

		// Collect receiver trackfiles: friendlies inside the hostile's threat radius that are on
		// frequency with the controller. These are the pilots who will actually hear the threat
		// call.
		receivers := make([]*trackfiles.Trackfile, 0)
		for _, friendlyGroup := range friendlyGroups {
			distance := spatial.Distance(grp.point(), friendlyGroup.point(), r.withProjection())
			withinThreatRadius := r.isGroupWithinThreatRadius(grp, distance)
			hostileIsHelo := grp.category() == brevity.RotaryWing
			friendlyIsPlane := friendlyGroup.category() == brevity.FixedWing
			heloVersusPlane := hostileIsHelo && friendlyIsPlane
			if !withinThreatRadius || heloVersusPlane {
				continue
			}
			for _, id := range friendlyGroup.ObjectIDs() {
				trackfile, ok := r.contacts.getByID(id)
				if !ok {
					continue
				}
				isOnFrequency := r.srsClient != nil && r.srsClient.IsOnFrequency(trackfile.Contact.Name)
				if !isOnFrequency {
					continue
				}
				receivers = append(receivers, trackfile)
			}
		}
		if len(receivers) == 0 {
			continue
		}

		ids := make([]uint64, 0, len(receivers))
		for _, tf := range receivers {
			ids = append(ids, tf.Contact.ID)
		}
		threats[grp] = ids

		// Pick a call format that best serves the filtered receiver set:
		//   - 1 receiver                → BRAA from that receiver.
		//   - tightly-grouped receivers → BRAA from the geographic midpoint, usable by all of them.
		//   - otherwise                 → bullseye (default from enumerateGroups).
		if len(receivers) == 1 {
			r.setGroupBRAA(grp, receivers[0].LastKnown().Point)
			continue
		}
		if origin, ok := r.getGroupBRAAOrigin(grp, receivers); ok {
			r.setGroupBRAA(grp, origin)
		}
	}

	result := make(map[brevity.Group][]uint64)
	for grp, ids := range threats {
		result[grp] = ids
	}
	return result
}

// setGroupBRAA populates the hostile group's BRAA relative to the given origin point and
// clears its bullseye so the THREAT call renders as BRAA.
func (r *Radar) setGroupBRAA(grp *group, origin orb.Point) {
	declination := r.Declination(origin)
	bearing := spatial.TrueBearing(origin, grp.point(), r.withProjection()).Magnetic(declination)
	_range := spatial.Distance(origin, grp.point(), r.withProjection())
	aspect := brevity.UnknownAspect
	if course, ok := grp.course(); ok {
		aspect = brevity.AspectFromAngle(bearing, course)
	}
	grp.braa = brevity.NewBRAA(bearing, _range, grp.altitudes(), aspect)
	grp.bullseye = nil
}

// getGroupBRAAOrigin returns the geographic midpoint of the receivers' positions if their
// BRAAs to the hostile are tightly enough grouped that a BRAA from the midpoint is within an
// acceptable error bound of each receiver's own BRAA. Otherwise it returns false, signalling
// that a bullseye call is more appropriate.
func (r *Radar) getGroupBRAAOrigin(hostile *group, receivers []*trackfiles.Trackfile) (orb.Point, bool) {
	if len(receivers) < 2 {
		return orb.Point{}, false
	}
	hostilePoint := hostile.point()
	if r.bearingSpread(hostilePoint, receivers) > r.maxSharedBRAABearingSpread {
		return orb.Point{}, false
	}
	if rangeSpread(hostilePoint, receivers, r.withProjection()) > r.maxSharedBRAARangeSpread {
		return orb.Point{}, false
	}
	return midpoint(receivers), true
}

// bearingSpread returns the widest magnetic bearing spread between any two receivers'
// BRAAs to the hostile.
func (r *Radar) bearingSpread(hostile orb.Point, receivers []*trackfiles.Trackfile) unit.Angle {
	bearingsToHostile := make([]bearings.Bearing, 0, len(receivers))
	for _, tf := range receivers {
		receiver := tf.LastKnown().Point
		declination := r.Declination(receiver)
		bearing := spatial.TrueBearing(receiver, hostile, r.withProjection()).Magnetic(declination)
		bearingsToHostile = append(bearingsToHostile, bearing)
	}
	widest := unit.Angle(0)
	for i, a := range bearingsToHostile {
		for _, b := range bearingsToHostile[i+1:] {
			if d := bearings.AngularDistance(a, b); d > widest {
				widest = d
			}
		}
	}
	return widest
}

// rangeSpread returns the difference between the longest and shortest range among the
// receivers' BRAAs to the hostile.
func rangeSpread(hostile orb.Point, receivers []*trackfiles.Trackfile, opt spatial.Option) unit.Length {
	minRange := unit.Length(math.Inf(1))
	maxRange := unit.Length(math.Inf(-1))
	for _, tf := range receivers {
		r := spatial.Distance(tf.LastKnown().Point, hostile, opt)
		if r < minRange {
			minRange = r
		}
		if r > maxRange {
			maxRange = r
		}
	}
	return maxRange - minRange
}

// midpoint of the given trackfiles.
func midpoint(contacts []*trackfiles.Trackfile) orb.Point {
	p := contacts[0].LastKnown().Point
	for _, tf := range contacts[1:] {
		p = geo.Midpoint(p, tf.LastKnown().Point)
	}
	return p
}

func (r *Radar) isGroupWithinThreatRadius(grp *group, distance unit.Length) bool {
	if !grp.isArmed() {
		return false
	}
	return distance < grp.threatRadius() || distance < r.mandatoryThreatRadius
}
