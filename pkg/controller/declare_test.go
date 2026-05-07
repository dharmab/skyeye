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

func TestHandleDeclare_CallsignNotOnRadar(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    30 * unit.NauticalMile,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.NegativeRadarContactResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
}

func TestHandleDeclare_Sour(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		Sour:     true,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.True(t, resp.Sour)
	assert.Equal(t, brevity.Unable, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleDeclare_BRAA_Clean(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	// Point BRAA at empty space far east — no aircraft there
	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    80 * unit.NauticalMile,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Clean, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleDeclare_BRAA_Friendly(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})

	// BRAA pointing roughly at Eagle 2 (~25nm east)
	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    25 * unit.NauticalMile,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Friendly, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Friendly, resp.Group.Declaration())
}

func TestHandleDeclare_BRAA_Hostile(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    25 * unit.NauticalMile,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Hostile, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, 1, resp.Group.Contacts())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleDeclare_BRAA_Furball(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place friendly and hostile within 7nm of declared point
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.01})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    25 * unit.NauticalMile,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Furball, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Furball, resp.Group.Declaration())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleDeclare_Bullseye_Hostile(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	// Place hostile at offset from bullseye (bullseye is at 30.0, 40.0)
	// ~25nm east of bullseye
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   false,
		Bullseye: brevity.NewBullseye(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile),
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Hostile, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleDeclare_Bullseye_NilBullseye(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   false,
		Bullseye: nil,
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.Equal(t, brevity.Unable, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleDeclare_Bullseye_Clean(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

	// Point bullseye at empty space far east — no aircraft there
	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   false,
		Bullseye: brevity.NewBullseye(bearings.NewMagneticBearing(90*unit.Degree), 80*unit.NauticalMile),
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Clean, resp.Declaration)
	assert.Nil(t, resp.Group)
}

func TestHandleDeclare_Bullseye_Friendly(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   false,
		Bullseye: brevity.NewBullseye(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile),
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Friendly, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Friendly, resp.Group.Declaration())
}

func TestHandleDeclare_Bullseye_Furball(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Eagle 2 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.5, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.01})

	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   false,
		Bullseye: brevity.NewBullseye(bearings.NewMagneticBearing(90*unit.Degree), 25*unit.NauticalMile),
		Altitude: 20000 * unit.Foot,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Furball, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Furball, resp.Group.Declaration())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}

func TestHandleDeclare_BRAA_ZeroAltitude(t *testing.T) {
	t.Parallel()
	h := newControllerTestHarness(t, nil)
	h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})
	h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

	// Altitude=0 should use full altitude range and still find the hostile
	h.ctrl.HandleDeclare(h.ctx, &brevity.DeclareRequest{
		Callsign: "eagle 1",
		IsBRAA:   true,
		Bearing:  bearings.NewMagneticBearing(90 * unit.Degree),
		Range:    25 * unit.NauticalMile,
		Altitude: 0,
	})
	got := h.expectResponse(t)
	resp, ok := got.(brevity.DeclareResponse)
	require.True(t, ok)
	assert.Equal(t, "eagle 1", resp.Callsign)
	assert.False(t, resp.Sour)
	assert.Equal(t, brevity.Hostile, resp.Declaration)
	require.NotNil(t, resp.Group)
	assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
	assert.Equal(t, 1, resp.Group.Contacts())
	assert.Contains(t, resp.Group.Platforms(), "Flanker")
}
