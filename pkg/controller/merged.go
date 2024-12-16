package controller

import (
	"context"
	"maps"
	"slices"
	"sync"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

// mergeTracker tracks hostile IDs and the friendly IDs they have merged with.
type mergeTracker struct {
	merged map[uint64]map[uint64]struct{}
	lock   sync.RWMutex
}

func newMergeTracker() *mergeTracker {
	return &mergeTracker{
		merged: make(map[uint64]map[uint64]struct{}),
	}
}

// merge records that the given hostile has merged with the given friendly.
func (t *mergeTracker) merge(hostileID, friendID uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	friendIDs, ok := t.merged[hostileID]
	if !ok {
		friendIDs = make(map[uint64]struct{})
		t.merged[hostileID] = friendIDs
	}
	friendIDs[friendID] = struct{}{}
}

// isMerged checks if the given hostile has merged with the given friendly.
func (t *mergeTracker) isMerged(hostileID, friendID uint64) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	friendIDs, ok := t.merged[hostileID]
	if !ok {
		return false
	}
	_, ok = friendIDs[friendID]
	return ok
}

// friendliesMergedWith returns the unique friendly IDs that have merged with the given hostile IDs.
func (t *mergeTracker) friendliesMergedWith(hostileIDs ...uint64) []uint64 {
	t.lock.RLock()
	defer t.lock.RUnlock()

	uniqueFriendIDs := make(map[uint64]struct{})
	for _, hostileID := range hostileIDs {
		if friendIDs, ok := t.merged[hostileID]; ok {
			for id := range friendIDs {
				uniqueFriendIDs[id] = struct{}{}
			}
		}
	}
	return slices.Collect(maps.Keys(uniqueFriendIDs))
}

// separate records that the given hostile and friendly IDs have exited the merge.
func (t *mergeTracker) separate(hostileID, friendID uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	friendIDs, ok := t.merged[hostileID]
	if !ok {
		return
	}
	delete(friendIDs, friendID)
	if len(friendIDs) == 0 {
		delete(t.merged, hostileID)
	}
}

// remove removes the given ID from the merge tracker.
func (t *mergeTracker) remove(id uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	_, ok := t.merged[id]
	if ok {
		delete(t.merged, id)
	} else {
		for hostileID, friendIDs := range t.merged {
			delete(friendIDs, id)
			if len(friendIDs) == 0 {
				delete(t.merged, hostileID)
			}
		}
	}
}

// keep removes any IDs that are not in the given slice from the merge tracker.
func (t *mergeTracker) keep(idsToKeep ...uint64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	for id := range t.merged {
		if !slices.Contains(idsToKeep, id) {
			delete(t.merged, id)
		}
	}
}

func (t *mergeTracker) reset() {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.merged = make(map[uint64]map[uint64]struct{})
}

// broadcastMerges updates the merge tracker and broadcasts merged calls for any new merges.
func (c *Controller) broadcastMerges(ctx context.Context) {
	merges := c.scope.Merges(c.coalition)

	hostileIDs := make([]uint64, 0)
	for group := range merges {
		hostileIDs = append(hostileIDs, group.ObjectIDs()...)
	}
	c.merges.keep(hostileIDs...)

	for hostileGroup, friendlies := range merges {
		logger := log.With().Stringer("group", hostileGroup).Logger()
		newMergedFriendlies := c.updateMergesForGroup(hostileGroup, friendlies)
		friendliesToNotify := make([]*trackfiles.Trackfile, 0)
		for _, friendly := range newMergedFriendlies {
			if c.mergeCooldowns.isOnCooldown(friendly.Contact.ID) {
				logger.Info().Uint64("friendID", friendly.Contact.ID).Msg("removing friend from pending merged call because another merged call was recently broadcast for this friend")
				continue
			}
			friendliesToNotify = append(friendliesToNotify, friendly)
			c.mergeCooldowns.extendCooldown(friendly.Contact.ID)
		}

		mergedCall := c.createMergedCall(friendliesToNotify)
		if len(mergedCall.Callsigns) > 0 {
			logger.Info().Strs("callsigns", mergedCall.Callsigns).Msg("broadcasting merged call")
			c.calls <- NewCall(ctx, mergedCall)
		} else {
			logger.Debug().Msg("skipping merged call because no relevant clients are on frequency")
		}
	}
}

// updateMergesForGroup updates the merge tracker for the given hostile group and friendly contacts.
// Friendlies which are newly merged with the hostile group are returned.
func (c *Controller) updateMergesForGroup(hostileGroup brevity.Group, friendlies []*trackfiles.Trackfile) []*trackfiles.Trackfile {
	friendIDs := make(map[uint64]struct{})
	for _, friendly := range friendlies {
		friendIDs[friendly.Contact.ID] = struct{}{}
	}

	newMergedFriendlies := make([]*trackfiles.Trackfile, 0)
	for _, hostileID := range hostileGroup.ObjectIDs() {
		for _, oldMergedFriendly := range c.merges.friendliesMergedWith(hostileID) {
			if _, ok := friendIDs[oldMergedFriendly]; !ok {
				c.merges.separate(hostileID, oldMergedFriendly)
			}
		}

		hostile := c.scope.FindUnit(hostileID)
		if hostile == nil {
			c.merges.remove(hostileID)
			continue
		}

		for _, friendly := range friendlies {
			isNewMerge := c.updateMergesForContact(hostile, friendly)
			if isNewMerge {
				newMergedFriendlies = append(newMergedFriendlies, friendly)
			}
		}
	}
	return newMergedFriendlies
}

// updateMergesForContact checks if the given hostile and friendly have merged or separated, and updates the merge tracker accordingly.
// It returns true if the contacts were merged, or false if they were already merged or if they were separated.
func (c *Controller) updateMergesForContact(hostile, friendly *trackfiles.Trackfile) bool {
	logger := log.
		With().
		Str("hostile", hostile.Contact.Name).
		Uint64("hostileID", hostile.Contact.ID).
		Str("friendly", friendly.Contact.Name).
		Uint64("friendID", friendly.Contact.ID).
		Logger()

	isMerged := c.merges.isMerged(hostile.Contact.ID, friendly.Contact.ID)
	if friendly.IsLastKnownPointZero() || hostile.IsLastKnownPointZero() {
		c.merges.separate(hostile.Contact.ID, friendly.Contact.ID)
		return false
	}
	distance := spatial.Distance(friendly.LastKnown().Point, hostile.LastKnown().Point)
	enteredMerge := distance < brevity.MergeEntryDistance
	exitedMerge := distance > brevity.MergeExitDistance

	if !isMerged && enteredMerge {
		logger.Info().Msg("hostile and friendly merged")
		c.merges.merge(hostile.Contact.ID, friendly.Contact.ID)
		return true
	} else if isMerged && exitedMerge {
		logger.Info().Msg("hostile and friendly exited merge")
		c.merges.separate(hostile.Contact.ID, friendly.Contact.ID)
	} else if isMerged {
		logger.Debug().Msg("hostile and friendly were already merged")
	}
	return false
}

func (c *Controller) createMergedCall(friendlies []*trackfiles.Trackfile) brevity.MergedCall {
	call := brevity.MergedCall{
		Callsigns: make([]string, 0),
	}
	for _, friendly := range friendlies {
		call.Callsigns = c.addFriendlyToBroadcast(call.Callsigns, friendly)
	}
	return call
}

// fillInMergeDetails sets the group's merged-with count, and if it is greater than 0, declares the group to be a FURBALL.
func (c *Controller) fillInMergeDetails(group brevity.Group) {
	count := len(c.merges.friendliesMergedWith(group.ObjectIDs()...))
	group.SetMergedWith(count)
	if group.MergedWith() > 0 {
		group.SetDeclaration(brevity.Furball)
	}
}

func (c *Controller) isGroupMergedWithFriendly(hostileGroup brevity.Group, friendID uint64) bool {
	for _, hostileID := range hostileGroup.ObjectIDs() {
		if c.merges.isMerged(hostileID, friendID) {
			return true
		}
	}
	return false
}
