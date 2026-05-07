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
	resp, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
}

func TestHandleBogeyDope_NoHostiles(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{Callsign: "eagle 1", Filter: brevity.Aircraft})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.BogeyDopeResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
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
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	assert.Nil(t, resp.Group.Bullseye())
	require.NotNil(t, resp.Group.BRAA())
	assert.InDelta(t, 84.0, resp.Group.BRAA().Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 23.0, resp.Group.BRAA().Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.Group.BRAA().Altitude().Feet(), altitudeDeltaFeet)
	assert.Equal(t, brevity.Aspect(brevity.Drag), resp.Group.BRAA().Aspect())
	assert.Equal(t, brevity.East, resp.Group.Track())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
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
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	require.NotNil(t, resp.Group.BRAA())
	// The Su-27 is fixed-wing; the Ka-50 is rotary. With FixedWing filter,
	// the nearest match should be the Su-27 (farther away), not the Ka-50.
	assert.InDelta(t, 84.0, resp.Group.BRAA().Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 23.0, resp.Group.BRAA().Range().NauticalMiles(), rangeDeltaNauticalMiles)
	assert.InDelta(t, 20000.0, resp.Group.BRAA().Altitude().Feet(), altitudeDeltaFeet)
	assert.Equal(t, brevity.Aspect(brevity.Drag), resp.Group.BRAA().Aspect())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}
