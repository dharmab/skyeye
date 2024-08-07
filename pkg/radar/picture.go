package radar

import (
	"math"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// GetPicture implements [Radar.GetPicture]
func (s *scope) GetPicture(radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) (int, []brevity.Group) {
	origin := s.center
	if origin.Lon() == 0 && origin.Lat() == 0 {
		log.Warn().Msg("center point is not set yet, using bullseye")
		origin = s.Bullseye(coalition)
		if origin.Lon() == 0 && origin.Lat() == 0 {
			log.Warn().Msg("bullseye point is not yet set, picture will be incoherent")
		}
	}

	visitedContacts := make(map[uint32]bool)
	groups := make([]*group, 0)
	itr := s.contacts.itr()
	for itr.next() {
		trackfile := itr.value()
		logger := log.With().Int("unitID", int(trackfile.Contact.UnitID)).Logger()
		if _, ok := visitedContacts[trackfile.Contact.UnitID]; ok {
			logger.Trace().Msg("skipping visited contact")
			continue
		}
		visitedContacts[trackfile.Contact.UnitID] = true
		if trackfile.Contact.Coalition != coalition {
			logger.Trace().Msg("skipping contact from other coalition")
			continue
		}
		if !isValidTrack(trackfile) {
			logger.Trace().Msg("skipping invalid track")
			continue
		}
		distance := unit.Length(geo.Distance(origin, trackfile.LastKnown().Point)) * unit.Meter
		if distance > radius {
			logger.Debug().Float64("distanceNM", distance.NauticalMiles()).Float64("radiusNM", radius.NauticalMiles()).Msg("skipping contact outside radius")
			continue
		}

		grp := s.findGroupForAircraft(trackfile)
		if grp == nil {
			logger.Error().Msg("failed to find group for aircraft - HOW DID YOU GET HERE")
			continue
		}
		logger = logger.With().Str("group", grp.String()).Logger()
		for _, contact := range grp.contacts {
			visitedContacts[contact.Contact.UnitID] = true
		}
		logger.Debug().Msg("accounted group")
		groups = append(groups, grp)
	}

	// Sort groups from highest to lowest threat
	slices.SortFunc(groups, s.compareThreat)

	// Return the top 3 groups
	capacity := 3
	if len(groups) < capacity {
		capacity = len(groups)
	}
	result := make([]brevity.Group, capacity)
	for i := 0; i < capacity; i++ {
		result[i] = groups[i]
	}
	return len(groups), result
}

func (s *scope) compareThreat(a, b *group) int {
	aIsHigherThreat := -1
	bIsHigherThreat := 1

	// Priotize armed aircraft over unarmed aircraft
	aIsArmed := a.threatFactor() > 0
	bIsArmed := b.threatFactor() > 0
	if aIsArmed && !bIsArmed {
		return aIsHigherThreat
	} else if !aIsArmed && bIsArmed {
		return bIsHigherThreat
	}

	// Prioritize fixed-wing aircraft over rotary-wing aircraft
	aIsHelo := a.category() == brevity.RotaryWing
	bIsHelo := b.category() == brevity.RotaryWing
	if !aIsHelo && bIsHelo {
		return aIsHigherThreat
	} else if aIsHelo && !bIsHelo {
		return bIsHigherThreat
	}

	// Remaining factors - distance, altitude, general threat factor, and number of contacts - are used as soft factors
	// ACC says there's an order of priority, but testing revealed that strict ordering resulted in closer Frogfoots being
	// ordered before somewhat further Flankers.
	//
	// This formula isn't scientific, I just tried some parameters until I got something that seemed to work okay.
	// TODO come back and do a proper formula.
	weight := func(g *group) float64 {
		return g.distanceWeight(s.center) * g.threatFactor() * g.altitudeWeight() * float64(g.Contacts())
	}
	if weight(a) > weight(b) {
		return aIsHigherThreat
	} else {
		return bIsHigherThreat
	}

}

func (g *group) distanceWeight(origin orb.Point) float64 {
	distance := unit.Length(geo.Distance(origin, g.point())) * unit.Meter
	cutoff := 100 * unit.NauticalMile
	if distance > cutoff {
		distance = cutoff - 1*unit.Meter
	}
	return math.Pow((cutoff-distance).NauticalMiles()/10, 2)
}

func (grp *group) altitudeWeight() float64 {
	altitude := grp.Altitude()
	cutoff := 40000 * unit.Foot
	if altitude > cutoff {
		altitude = cutoff
	}
	return math.Pow((altitude.Feet()/10000)+1, 1.4)
}
