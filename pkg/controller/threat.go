package controller

import (
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/parser"
)

type cooldownTracker struct {
	// cooldowns is keyed by the unit IDs of coalition aircraft. The values are sub-maps keyed by the unit IDs of
	// hostile aircraft. The values of these sub-maps are the cooldown times after which the controller can issue
	// another threat call to the coalition aircraft in regards to the hostile aircraft.
	cooldowns map[uint32]map[uint32]time.Time
	// cooldownsLock is used to synchronize access to the cooldowns map.
	cooldownsLock sync.RWMutex
}

func newCooldownTracker() *cooldownTracker {
	return &cooldownTracker{
		cooldowns: make(map[uint32]map[uint32]time.Time),
	}
}

func (d *cooldownTracker) extendCooldown(subjectID uint32, objectIDs ...uint32) {
	d.cooldownsLock.Lock()
	defer d.cooldownsLock.Unlock()

	if _, ok := d.cooldowns[subjectID]; !ok {
		d.cooldowns[subjectID] = make(map[uint32]time.Time)
	}
	for _, objectID := range objectIDs {
		d.cooldowns[subjectID][objectID] = time.Now().Add(5 * time.Minute)
	}
}

func (d *cooldownTracker) isOnCooldown(subjectID, objectID uint32) bool {
	d.cooldownsLock.RLock()
	defer d.cooldownsLock.RUnlock()

	cooldowns, ok := d.cooldowns[subjectID]
	if !ok {
		return false
	}
	cooldown, ok := cooldowns[objectID]
	if !ok {
		return false
	}

	return time.Now().Before(cooldown)
}

func (d *cooldownTracker) remove(unitID uint32) {
	d.cooldownsLock.Lock()
	defer d.cooldownsLock.Unlock()
	for _, cooldowns := range d.cooldowns {
		delete(cooldowns, unitID)
	}
	delete(d.cooldowns, unitID)
}

func (c *controller) broadcastThreats() {
	threats := c.scope.Threats(c.hostileCoalition())
	for unitID, groups := range threats {
		if !c.srsClient.IsOnFrequency(unitID) {
			continue
		}

		for _, group := range groups {
			recentlyNotified := true
			for threatID := range group.UnitIDs() {
				if !c.threatCooldowns.isOnCooldown(unitID, uint32(threatID)) {
					recentlyNotified = false
					break
				}
			}
			if recentlyNotified {
				continue
			}
			trackfile := c.scope.FindUnit(unitID)
			if trackfile == nil {
				continue
			}
			callsign, ok := parser.ParsePilotCallsign(trackfile.Contact.ACMIName)
			if !ok {
				continue
			}

			c.threatCooldowns.extendCooldown(unitID, group.UnitIDs()...)

			if time.Now().Before(c.warmupTime) {
				continue
			}

			group.SetDeclaration(brevity.Hostile)
			group.SetThreat(true)
			c.out <- brevity.ThreatCall{Callsign: callsign, Group: group}
		}
	}
}
