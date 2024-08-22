package controller

import (
	"slices"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/rs/zerolog/log"
)

type cooldownTracker struct {
	cooldown time.Duration
	// cooldowns maps unit IDs to the time at which the the threat cooldown expires. The threat cooldown suppresses
	// threat calls for the threat with the given unit ID.
	cooldowns map[uint32]time.Time
	// lock used to synchronize access to the cooldowns map.
	lock sync.RWMutex
}

func newCooldownTracker(cooldown time.Duration) *cooldownTracker {
	return &cooldownTracker{
		cooldown:  cooldown,
		cooldowns: make(map[uint32]time.Time),
	}
}

func (t *cooldownTracker) extendCooldown(unitID uint32) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cooldowns[unitID] = time.Now().Add(t.cooldown)
}

func (t *cooldownTracker) isOnCooldown(unitID uint32) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	cooldown, ok := t.cooldowns[unitID]
	if !ok {
		return false
	}

	return time.Now().Before(cooldown)
}

func (t *cooldownTracker) remove(unitID uint32) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.cooldowns, unitID)
}

func (t *controller) broadcastThreats() {
	if !t.enableThreatMonitoring {
		return
	}

	threats := t.scope.Threats(t.coalition.Opposite())
	for group, unitIDs := range threats {
		logger := log.With().Stringer("group", group).Uints32("unitIDs", unitIDs).Logger()
		isOnFrequency := false
		for _, unitID := range unitIDs {
			if t.srsClient.IsOnFrequency(unitID) {
				isOnFrequency = true
			}
			// if only one unit of multiple is on frequency, we still use bullsye instead
			// of BRAA. can we do better?
		}

		if t.threatMonitoringRequiresSRS && !isOnFrequency {
			logger.Info().Uints32("unitIDs", unitIDs).Msg("supressing threat call because units are not on frequency")
			continue
		}

		recentlyNotified := true
		for threatID := range group.UnitIDs() {
			if !t.threatCooldowns.isOnCooldown(uint32(threatID)) {
				recentlyNotified = false
				break
			}
		}
		if recentlyNotified {
			logger.Info().Uints32("threatIDs", group.UnitIDs()).Msg("supressing threat call because a call was recently broadcast for this threat")
			continue
		}

		for threatID := range group.UnitIDs() {
			t.threatCooldowns.extendCooldown(uint32(threatID))
		}

		group.SetDeclaration(brevity.Hostile)
		group.SetThreat(true)
		call := brevity.ThreatCall{Group: group}

		for _, unitID := range unitIDs {
			if trackfile := t.scope.FindUnit(unitID); trackfile != nil {
				if callsign, ok := parser.ParsePilotCallsign(trackfile.Contact.ACMIName); ok {
					if !slices.Contains(call.Callsigns, callsign) {
						call.Callsigns = append(call.Callsigns, callsign)
					}
				}
			}
		}

		logger.Info().Any("call", call).Msg("broadcasting threat call for group")
		t.out <- call
	}
}
