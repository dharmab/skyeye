package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUnableToUnderstand_CallsignOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.1, 40.1})

	h.ctrl.HandleUnableToUnderstand(h.ctx, &brevity.UnableToUnderstandRequest{Callsign: "eagle 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SayAgainResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
}

func TestHandleUnableToUnderstand_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleUnableToUnderstand(h.ctx, &brevity.UnableToUnderstandRequest{Callsign: "ghost 1"})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SayAgainResponse)
	require.True(t, ok)
	assert.Equal(t, brevity.LastCaller, resp.Callsign)
}
