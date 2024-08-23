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
	cooldowns map[uint64]time.Time
	// lock used to synchronize access to the cooldowns map.
	lock sync.RWMutex
}

func newCooldownTracker(cooldown time.Duration) *cooldownTracker {
	return &cooldownTracker{
		cooldown:  cooldown,
		cooldowns: make(map[uint64]time.Time),
	}
}

func (t *cooldownTracker) extendCooldown(id uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cooldowns[id] = time.Now().Add(t.cooldown)
}

func (t *cooldownTracker) isOnCooldown(id uint64) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	cooldown, ok := t.cooldowns[id]
	if !ok {
		return false
	}

	return time.Now().Before(cooldown)
}

func (t *cooldownTracker) remove(id uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.cooldowns, id)
}

func (t *controller) broadcastThreats() {
	if !t.enableThreatMonitoring {
		return
	}

	threats := t.scope.Threats(t.coalition.Opposite())
	for group, ids := range threats {
		logger := log.With().Stringer("group", group).Uints64("ids", ids).Logger()

		recentlyNotified := true
		for _, threatID := range group.ObjectIDs() {
			if !t.threatCooldowns.isOnCooldown(threatID) {
				recentlyNotified = false
				break
			}
		}
		if recentlyNotified {
			logger.Debug().Uints64("threatIDs", group.ObjectIDs()).Msg("supressing threat call because a call was recently broadcast for this threat")
			continue
		}

		group.SetDeclaration(brevity.Hostile)
		group.SetThreat(true)
		call := brevity.ThreatCall{Group: group}

		for _, id := range ids {
			if trackfile := t.scope.FindUnit(id); trackfile != nil {
				isOnFrequency := t.srsClient.IsOnFrequency(trackfile.Contact.Name)
				if !t.threatMonitoringRequiresSRS || isOnFrequency {
					if callsign, ok := parser.ParsePilotCallsign(trackfile.Contact.Name); ok {
						if !slices.Contains(call.Callsigns, callsign) {
							call.Callsigns = append(call.Callsigns, callsign)
						}
					}
				}
			}
		}

		if len(call.Callsigns) == 0 {
			logger.Debug().Msg("skipping threat call because there is no one to notify")
			continue
		}

		logger.Info().Any("call", call).Msg("broadcasting threat call for group")
		t.out <- call

		for _, threatID := range group.ObjectIDs() {
			t.threatCooldowns.extendCooldown(threatID)
		}
	}
}
