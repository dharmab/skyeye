package controller

import (
	"context"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

type cooldownTracker struct {
	// cooldown is the interval between threat calls for the same threat.
	cooldown time.Duration
	// cooldowns maps unit IDs to the time at which the threat cooldown expires. The threat cooldown suppresses
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

func (t *cooldownTracker) reset() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.cooldowns = make(map[uint64]time.Time)
}

func (t *cooldownTracker) remove(id uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.cooldowns, id)
}

func (c *Controller) broadcastThreats(ctx context.Context) {
	if !c.enableThreatMonitoring {
		return
	}
	threats := c.scope.Threats(c.coalition.Opposite())
	for hostileGroup, friendIDs := range threats {
		c.broadcastThreat(ctx, hostileGroup, friendIDs)
	}
}

func (c *Controller) broadcastThreat(ctx context.Context, hostileGroup brevity.Group, friendIDs []uint64) {
	hostileGroup.SetDeclaration(brevity.Hostile)
	c.fillInMergeDetails(hostileGroup)
	hostileGroup.SetThreat(true)

	logger := log.With().Stringer("group", hostileGroup).Uints64("friendIDs", friendIDs).Logger()

	recentlyNotified := true
	for _, threatID := range hostileGroup.ObjectIDs() {
		if !c.threatCooldowns.isOnCooldown(threatID) {
			recentlyNotified = false
			break
		}
	}
	if recentlyNotified {
		logger.Debug().Uints64("threatIDs", hostileGroup.ObjectIDs()).Msg("suppressing threat call because a call was recently broadcast for all contacts within the threat group")
		return
	}

	threatCall := brevity.ThreatCall{
		Callsigns: make([]string, 0),
		Group:     hostileGroup,
	}

	for _, friendID := range friendIDs {
		if c.isGroupMergedWithFriendly(hostileGroup, friendID) {
			logger.Debug().Msg("omitting friendly from threat call because the threat is already merged")
			continue
		}
		if friendly := c.scope.FindUnit(friendID); friendly != nil {
			threatCall.Callsigns = c.addFriendlyToBroadcast(threatCall.Callsigns, friendly)
		}
	}

	if len(threatCall.Callsigns) == 0 {
		logger.Debug().Msg("skipping threat call because no relevant clients are on frequency")
		return
	}

	logger.Info().Any("call", threatCall).Msg("broadcasting threat call for group")
	c.calls <- NewCall(ctx, threatCall)

	for _, threatID := range hostileGroup.ObjectIDs() {
		c.threatCooldowns.extendCooldown(threatID)
	}
}
