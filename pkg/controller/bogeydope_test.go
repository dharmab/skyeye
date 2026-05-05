package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleBogeyDope_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{Callsign: "eagle 1", Filter: brevity.Aircraft})
	got := h.expectResponse(t)
	_, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok, "got %T", got)
}

func TestHandleBogeyDope_NoHostiles(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{Callsign: "eagle 1", Filter: brevity.Aircraft})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.BogeyDopeResponse)
	require.True(t, ok, "got %T", got)
	assert.Nil(t, resp.Group)
}

func TestHandleBogeyDope_HostilePresent(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{Callsign: "eagle 1", Filter: brevity.Aircraft})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.BogeyDopeResponse)
	require.True(t, ok, "got %T", got)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	require.NotNil(t, resp.Group.BRAA())
	assert.InDelta(t, 25.0, resp.Group.BRAA().Range().NauticalMiles(), 10.0)
}

func TestHandleBogeyDope_FilterFixedWing(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Bandit Helo", acmiKa50, coalitions.Red, orb.Point{30.1, 40.0})
	h.insertAircraft(t, "Bandit Fighter", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{Callsign: "eagle 1", Filter: brevity.FixedWing})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.BogeyDopeResponse)
	require.True(t, ok, "got %T", got)
	require.NotNil(t, resp.Group)
	// The Su-27 is fixed-wing; the Ka-50 is rotary. With FixedWing filter,
	// the nearest match should be the Su-27 (farther away), not the Ka-50.
	assert.InDelta(t, 25.0, resp.Group.BRAA().Range().NauticalMiles(), 10.0)
}
