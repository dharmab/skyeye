package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleAlphaCheck_CallsignOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleAlphaCheck(h.ctx, &brevity.AlphaCheckRequest{Callsign: "eagle 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.AlphaCheckResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Location)
	assert.InDelta(t, 31.0, resp.Location.Bearing().Degrees(), bearingDeltaDegrees)
	assert.InDelta(t, 8.0, resp.Location.Distance().NauticalMiles(), rangeDeltaNauticalMiles)
}

func TestHandleAlphaCheck_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleAlphaCheck(h.ctx, &brevity.AlphaCheckRequest{Callsign: "eagle 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.AlphaCheckResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Status)
	assert.Nil(t, resp.Location)
}
