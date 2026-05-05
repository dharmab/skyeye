package controller

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSpiked_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleSpiked(h.ctx, &brevity.SpikedRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	_, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok, "got %T", got)
}

func TestHandleSpiked_NoHostileInCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleSpiked(h.ctx, &brevity.SpikedRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SpikedResponseV2)
	require.True(t, ok, "got %T", got)
	assert.False(t, resp.Status)
}

func TestHandleSpiked_HostileInCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place hostile roughly east (~25nm) — within 30-degree cone of magnetic bearing 090
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleSpiked(h.ctx, &brevity.SpikedRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SpikedResponseV2)
	require.True(t, ok, "got %T", got)
	assert.True(t, resp.Status)
	require.NotNil(t, resp.Group)
}

func TestHandleSpiked_HostileOutsideCone(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place hostile roughly north — well outside 30-degree cone of magnetic bearing 090
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.0, 40.5})

	h.ctrl.HandleSpiked(h.ctx, &brevity.SpikedRequest{
		Callsign: "eagle 1",
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.SpikedResponseV2)
	require.True(t, ok, "got %T", got)
	assert.False(t, resp.Status)
}
