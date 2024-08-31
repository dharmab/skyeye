package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeTrackerMerge(t *testing.T) {
	t.Parallel()
	tracker := newMergeTracker()
	tracker.merge(1, 2)
	assert.True(t, tracker.isMerged(1, 2))
	assert.False(t, tracker.isMerged(2, 1))
	assert.False(t, tracker.isMerged(1, 3))
	tracker.merge(1, 3)
	assert.True(t, tracker.isMerged(1, 3))
	tracker.merge(4, 1)
	assert.True(t, tracker.isMerged(4, 1))
}

func TestMergeTrackerFriendliesMergedWith(t *testing.T) {
	t.Parallel()
	tracker := newMergeTracker()
	tracker.merge(1, 2)
	assert.Len(t, tracker.friendliesMergedWith(1), 1)
	assert.Contains(t, tracker.friendliesMergedWith(1), uint64(2))
	assert.Empty(t, tracker.friendliesMergedWith(3))
	tracker.merge(1, 3)
	assert.Len(t, tracker.friendliesMergedWith(1), 2)
	assert.Contains(t, tracker.friendliesMergedWith(1), uint64(3))
}

func TestMergeTrackerSeparate(t *testing.T) {
	t.Parallel()
	tracker := newMergeTracker()
	tracker.merge(1, 2)
	tracker.merge(1, 3)
	assert.True(t, tracker.isMerged(1, 2))
	assert.True(t, tracker.isMerged(1, 3))
	tracker.separate(1, 2)
	assert.False(t, tracker.isMerged(1, 2))
	assert.True(t, tracker.isMerged(1, 3))
}

func TestMergeTrackerRemove(t *testing.T) {
	t.Parallel()
	tracker := newMergeTracker()
	red1 := uint64(1)
	red2 := uint64(2)
	red3 := uint64(3)
	blue1 := uint64(11)
	blue2 := uint64(12)
	blue3 := uint64(13)

	for _, hostile := range []uint64{red1, red2, red3} {
		for _, friendly := range []uint64{blue1, blue2, blue3} {
			tracker.merge(hostile, friendly)
		}
	}

	assert.True(t, tracker.isMerged(red1, blue1))
	assert.True(t, tracker.isMerged(red1, blue2))
	assert.True(t, tracker.isMerged(red1, blue3))
	assert.True(t, tracker.isMerged(red2, blue1))
	assert.True(t, tracker.isMerged(red2, blue2))
	assert.True(t, tracker.isMerged(red2, blue3))
	assert.True(t, tracker.isMerged(red3, blue1))
	assert.True(t, tracker.isMerged(red3, blue2))
	assert.True(t, tracker.isMerged(red3, blue3))

	// Remove a hostile ID
	tracker.remove(red1)
	assert.False(t, tracker.isMerged(red1, blue1))
	assert.False(t, tracker.isMerged(red1, blue2))
	assert.False(t, tracker.isMerged(red1, blue3))
	assert.True(t, tracker.isMerged(red2, blue1))
	assert.True(t, tracker.isMerged(red2, blue2))
	assert.True(t, tracker.isMerged(red2, blue3))
	assert.True(t, tracker.isMerged(red3, blue1))
	assert.True(t, tracker.isMerged(red3, blue2))
	assert.True(t, tracker.isMerged(red3, blue3))

	// Remove a friendly ID
	tracker.remove(blue1)
	assert.False(t, tracker.isMerged(red1, blue1))
	assert.False(t, tracker.isMerged(red1, blue2))
	assert.False(t, tracker.isMerged(red1, blue3))
	assert.False(t, tracker.isMerged(red2, blue1))
	assert.True(t, tracker.isMerged(red2, blue2))
	assert.True(t, tracker.isMerged(red2, blue3))
	assert.False(t, tracker.isMerged(red3, blue1))
	assert.True(t, tracker.isMerged(red3, blue2))
	assert.True(t, tracker.isMerged(red3, blue3))
}
func TestMergeTrackerKeep(t *testing.T) {
	t.Parallel()
	tracker := newMergeTracker()
	tracker.merge(1, 11)
	tracker.merge(1, 12)
	tracker.merge(2, 11)
	tracker.merge(3, 12)
	tracker.merge(4, 11)
	tracker.keep(2, 4)
	assert.False(t, tracker.isMerged(1, 11))
	assert.False(t, tracker.isMerged(1, 12))
	assert.True(t, tracker.isMerged(2, 11))
	assert.False(t, tracker.isMerged(3, 12))
	assert.True(t, tracker.isMerged(4, 11))
}
