package controller

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePicture_NoHostiles(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		h := newControllerTestHarness(t, nil)
		h.insertAircraft(t, "Eagle 1 Reaper", acmiF15C, coalitions.Blue, orb.Point{30.0, 40.0})

		time.Sleep(6 * time.Second)
		synctest.Wait()

		h.ctrl.HandlePicture(h.ctx, &brevity.PictureRequest{Callsign: "eagle 1"})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.PictureResponse)
		require.True(t, ok)
		assert.Equal(t, 0, resp.Count)
		assert.Empty(t, resp.Groups)
	})
}

func TestHandlePicture_OneHostileGroup(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		h := newControllerTestHarness(t, nil)
		h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})

		time.Sleep(6 * time.Second)
		synctest.Wait()

		h.ctrl.HandlePicture(h.ctx, &brevity.PictureRequest{Callsign: ""})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.PictureResponse)
		require.True(t, ok)
		assert.Equal(t, 1, resp.Count)
		require.Len(t, resp.Groups, 1)
		grp := resp.Groups[0]
		assert.Equal(t, brevity.Hostile, grp.Declaration())
		assert.Equal(t, 1, grp.Contacts())
		assert.Nil(t, grp.BRAA())
		require.NotNil(t, grp.Bullseye())
		assert.InDelta(t, 84.0, grp.Bullseye().Bearing().Degrees(), bearingDeltaDegrees)
		assert.InDelta(t, 23.0, grp.Bullseye().Distance().NauticalMiles(), rangeDeltaNauticalMiles)
		assert.Equal(t, brevity.East, grp.Track())
		assert.Contains(t, grp.Platforms(), "Flanker")
	})
}

func TestHandlePicture_MultipleHostileGroups(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		h := newControllerTestHarness(t, nil)
		// Place 3 hostile aircraft far enough apart to form separate groups (>3nm)
		h.insertAircraft(t, "Bandit 1", acmiSu27, coalitions.Red, orb.Point{30.5, 40.0})
		h.insertAircraft(t, "Bandit 2", acmiMiG29A, coalitions.Red, orb.Point{29.5, 40.0})
		h.insertAircraft(t, "Bandit 3", acmiSu27, coalitions.Red, orb.Point{30.0, 40.5})

		time.Sleep(6 * time.Second)
		synctest.Wait()

		h.ctrl.HandlePicture(h.ctx, &brevity.PictureRequest{Callsign: ""})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.PictureResponse)
		require.True(t, ok)
		assert.Equal(t, 3, resp.Count)
		require.Len(t, resp.Groups, 3)
		for _, grp := range resp.Groups {
			assert.Equal(t, brevity.Hostile, grp.Declaration())
			assert.Equal(t, 1, grp.Contacts())
			assert.Nil(t, grp.BRAA())
			require.NotNil(t, grp.Bullseye())
			assert.Equal(t, brevity.East, grp.Track())
		}
	})
}
