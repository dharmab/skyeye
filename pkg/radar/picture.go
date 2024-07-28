package radar

import (
	"math"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

func (s *scope) GetPicture(origin orb.Point, radius unit.Length, coalition coalitions.Coalition, filter brevity.ContactCategory) (int, []brevity.Group) {
	visitedContacts := make(map[uint32]bool)
	groups := make([]*group, 0)
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, tf := range s.contacts {
		logger := log.With().Int("unitID", int(tf.Contact.UnitID)).Logger()
		if visitedContacts[tf.Contact.UnitID] {
			logger.Trace().Msg("skipping visited contact")
			continue
		}
		visitedContacts[tf.Contact.UnitID] = true
		if tf.Contact.Coalition != coalition {
			logger.Trace().Msg("skipping contact from other coalition")
			continue
		}
		if !isValidTrack(tf) {
			logger.Trace().Msg("skipping invalid track")
			continue
		}
		distance := unit.Length(geo.Distance(origin, tf.LastKnown().Point)) * unit.Meter
		if distance > radius {
			logger.Debug().Float64("distanceNM", distance.NauticalMiles()).Float64("radiusNM", radius.NauticalMiles()).Msg("skipping contact outside radius")
			continue
		}

		group := s.findGroupForAircraft(tf)
		if group == nil {
			logger.Error().Msg("failed to find group for aircraft - HOW DID YOU GET HERE")
			continue
		}
		logger = logger.With().Any("group", group).Logger()
		for _, contact := range group.contacts {
			visitedContacts[contact.Contact.UnitID] = true
		}
		logger.Debug().Msg("accounted group")
		groups = append(groups, group)
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

		// Prioritize aircraft based on distance from caller
		// 3NM is considered a marginal difference
		distanceA := unit.Length(geo.Distance(origin, a.point())) * unit.Meter
		distanceB := unit.Length(geo.Distance(origin, b.point())) * unit.Meter
		isDistanceSimilar := math.Abs(
			distanceA.NauticalMiles()-distanceB.NauticalMiles(),
		) < 3
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
	for i := 0; i < len(groups) && i < 3; i++ {
		result[i] = groups[i]
	}
	return len(groups), result
}
