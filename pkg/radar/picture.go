package radar

import (
	"math"
	"slices"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/martinlindhe/unit"
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

	aIsHigherThreat := -1
	bIsHigherThreat := 1
	slices.SortFunc(groups, func(a, b *group) int {
		// Priotize armed aircraft over unarmed aircraft
		aIsArmed := a.threatClass() != encyclopedia.NoFactor
		bIsArmed := b.threatClass() != encyclopedia.NoFactor
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

		// Prioritize aircraft based on distance from origin
		distanceA := unit.Length(geo.Distance(origin, a.point())) * unit.Meter
		distanceB := unit.Length(geo.Distance(origin, b.point())) * unit.Meter
		isDistanceSimilar := math.Abs(
			distanceA.NauticalMiles()-distanceB.NauticalMiles(),
		) < conf.DefaultMarginRadius.NauticalMiles()
		if !isDistanceSimilar {
			return int(distanceA - distanceB)
		}

		// Prioritize groups with a higher threat class
		if a.threatClass() != b.threatClass() {
			return int(b.threatClass() - a.threatClass())
		}

		// Prioritize HIGH groups
		if a.High() && !b.High() {
			return aIsHigherThreat
		} else if !a.High() && b.High() {
			return bIsHigherThreat
		}

		// Prioritize larger groups
		return b.Contacts() - a.Contacts()
	})

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
