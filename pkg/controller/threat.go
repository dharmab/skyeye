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
	// cooldown is the interval between threat calls for the same threat.
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

func (c *controller) broadcastThreats() {
	if !c.enableThreatMonitoring {
		return
	}

	threats := c.scope.Threats(c.coalition.Opposite())
	for group, ids := range threats {
		group.SetDeclaration(brevity.Hostile)
		group.SetThreat(true)

		logger := log.With().Stringer("group", group).Uints64("ids", ids).Logger()

		recentlyNotified := true
		for _, threatID := range group.ObjectIDs() {
			if !c.threatCooldowns.isOnCooldown(threatID) {
				recentlyNotified = false
				break
			}
		}
		if recentlyNotified {
			logger.Debug().Uints64("threatIDs", group.ObjectIDs()).Msg("supressing threat call because a call was recently broadcast for this threat")
			continue
		}

		call := brevity.ThreatCall{Group: group}

		for _, id := range ids {
			if trackfile := c.scope.FindUnit(id); trackfile != nil {
				isOnFrequency := c.srsClient.IsOnFrequency(trackfile.Contact.Name)
				if !c.threatMonitoringRequiresSRS || isOnFrequency {
					if callsign, ok := parser.ParsePilotCallsign(trackfile.Contact.Name); ok {
						if !slices.Contains(call.Callsigns, callsign) {
							call.Callsigns = append(call.Callsigns, callsign)
						}
					}
				}
			}
		}

		if len(call.Callsigns) == 0 {
			logger.Debug().Msg("skipping threat call because no relevant clients are on frequency")
			continue
		}

		logger.Info().Any("call", call).Msg("broadcasting threat call for group")
		c.out <- call

		for _, threatID := range group.ObjectIDs() {
			c.threatCooldowns.extendCooldown(threatID)
		}
	}
}
