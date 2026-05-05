package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCheckIn_CallsignOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleCheckIn(h.ctx, &brevity.CheckInRequest{Callsign: "eagle 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.CheckInResponse)
	require.True(t, ok, "got %T", got)
	assert.Contains(t, resp.Callsign, "eagle 1")
}

func TestHandleCheckIn_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleCheckIn(h.ctx, &brevity.CheckInRequest{Callsign: "eagle 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok, "got %T", got)
	assert.Equal(t, "eagle 1", resp.Callsign)
}
