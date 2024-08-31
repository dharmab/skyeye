package controller

import (
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

// friendliesMergedWith returns the IDS that the given hostile ID is merged with.
func (t *mergeTracker) friendliesMergedWith(hostileID uint64) []uint64 {
	t.lock.RLock()
	defer t.lock.RUnlock()
	friendIDs, ok := t.merged[hostileID]
	if !ok {
		return []uint64{}
	}
	return slices.Collect(maps.Keys(friendIDs))
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

// broadcastMerges updates the merge tracker and broadcasts merged calls for any new merges.
func (c *controller) broadcastMerges() {
	merges := c.scope.Merges(c.coalition)

	hostileIDs := make([]uint64, 0)
	for group := range merges {
		hostileIDs = append(hostileIDs, group.ObjectIDs()...)
	}
	c.merges.keep(hostileIDs...)

	for hostileGroup, friendlies := range merges {
		newMergedFriendlies := c.updateMergesForGroup(hostileGroup, friendlies)

		logger := log.With().Stringer("group", hostileGroup).Logger()
		call := c.createMergedCall(hostileGroup, newMergedFriendlies)
		if len(call.Callsigns) > 0 {
			logger.Info().Strs("callsigns", call.Callsigns).Msg("broadcasting merged call")
			c.out <- call
		} else {
			logger.Debug().Msg("skipping merged call because no relevant clients are on frequency")
		}
	}
}

// updateMergesForGroup updates the merge tracker for the given hostile group and friendly contacts.
// Friendlies which are newly merged with the hostile group are returned.
func (c *controller) updateMergesForGroup(hostileGroup brevity.Group, friendlies []*trackfiles.Trackfile) []*trackfiles.Trackfile {
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
func (c *controller) updateMergesForContact(hostile, friendly *trackfiles.Trackfile) bool {
	logger := log.
		With().
		Str("hostile", hostile.Contact.Name).
		Uint64("hostileID", hostile.Contact.ID).
		Str("friendly", friendly.Contact.Name).
		Uint64("friendID", friendly.Contact.ID).
		Logger()

	isMerged := c.merges.isMerged(hostile.Contact.ID, friendly.Contact.ID)
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

func (c *controller) createMergedCall(hostileGroup brevity.Group, friendlies []*trackfiles.Trackfile) brevity.MergedCall {
	call := brevity.MergedCall{
		Group:     hostileGroup,
		Callsigns: make([]string, 0),
	}
	for _, friendly := range friendlies {
		call.Callsigns = c.addFriendlyToBroadcast(call.Callsigns, friendly)
	}
	return call
}

// fillInMergeDetails sets the group's merged-with count, and if it is greater than 0, declares the group to be a FURBALL.
func (c *controller) fillInMergeDetails(group brevity.Group) {
	mergedWith := 0
	for _, id := range group.ObjectIDs() {
		mergedWith += len(c.merges.friendliesMergedWith(id))
	}
	group.SetMergedWith(mergedWith)
	if group.MergedWith() > 0 {
		group.SetDeclaration(brevity.Furball)
	}
}

func (c *controller) isGroupMergedWithFriendly(hostileGroup brevity.Group, friendID uint64) bool {
	for _, hostileID := range hostileGroup.ObjectIDs() {
		if c.merges.isMerged(hostileID, friendID) {
			return true
		}
	}
	return false
}
