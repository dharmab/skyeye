package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

func TestGetByCallsign(t *testing.T) {
	d := newContactDatabase()
	trackfile := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	d.set(trackfile)

	name, tf, ok := d.getByCallsignAndCoalititon("mobius 1", coalitions.Blue)
	require.True(t, ok)
	require.Equal(t, "mobius 1", name)
	require.EqualValues(t, trackfile, tf)

	_, _, ok = d.getByCallsignAndCoalititon("mobius 1", coalitions.Red)
	require.False(t, ok)

	name, tf, ok = d.getByCallsignAndCoalititon("moebius 1", coalitions.Blue)
	require.True(t, ok)
	require.Equal(t, "mobius 1", name)
	require.EqualValues(t, trackfile, tf)

	_, _, ok = d.getByCallsignAndCoalititon("yellow 13", coalitions.Red)
	require.False(t, ok)
}

func TestRealCallsigns(t *testing.T) {
	// Callsigns collected from Discord
	testCases := []struct {
		Name    string
		heardAs string
	}{
		{Name: "Hussein 1-1 | SpyderF16", heardAs: "houston 1 1"},
		{Name: "Witch 1-1", heardAs: "which 1 1"},
		{Name: "Spare 15", heardAs: "spear 15"},
	}
	d := newContactDatabase()

	for i, test := range testCases {
		trackfile := &trackfiles.Trackfile{
			Contact: trackfiles.Labels{
				UnitID:    uint32(i),
				Name:      test.Name,
				Coalition: coalitions.Blue,
				ACMIName:  "F-15C",
			},
			Track: *deque.New[trackfiles.Frame](),
		}
		d.set(trackfile)
	}

	for i, test := range testCases {
		parsedCallsign, ok := parser.ParsePilotCallsign(test.Name)
		require.True(t, ok)
		foundCallsign, tf, ok := d.getByCallsignAndCoalititon(test.heardAs, coalitions.Blue)
		require.True(t, ok, "queried %s, expected %s, but result was %v", test.heardAs, test.Name, ok)
		require.Equal(t, parsedCallsign, foundCallsign)
		require.EqualValues(t, uint32(i), tf.Contact.UnitID)
	}
}

func TestGetByUnitID(t *testing.T) {
	d := newContactDatabase()
	trackfile := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	d.set(trackfile)

	val, ok := d.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)

	_, ok = d.getByUnitID(2)
	require.False(t, ok)
}

func TestSet(t *testing.T) {
	database := newContactDatabase()
	trackfile := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	database.set(trackfile)

	val, ok := database.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)

	trackfile.Update(trackfiles.Frame{
		Timestamp: time.Now(),
		Point: orb.Point{
			1,
			1,
		},
		Altitude: 1000 * unit.Foot,
		Heading:  90 * unit.Degree,
	})

	database.set(trackfile)

	val, ok = database.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)
}

func TestDelete(t *testing.T) {
	database := newContactDatabase()
	trackfile := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	database.set(trackfile)

	_, ok := database.getByUnitID(1)
	require.True(t, ok)

	ok = database.delete(1)
	require.True(t, ok)

	_, ok = database.getByUnitID(1)
	require.False(t, ok)

	ok = database.delete(2)
	require.False(t, ok)
}

func TestItr(t *testing.T) {
	database := newContactDatabase()
	mobius := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	database.set(mobius)

	yellow := &trackfiles.Trackfile{
		Contact: trackfiles.Labels{
			UnitID:    2,
			Name:      "Yellow 13 Reiher",
			Coalition: coalitions.Red,
			ACMIName:  "Su-27",
		},
		Track: *deque.New[trackfiles.Frame](),
	}
	database.set(yellow)

	itr := database.itr()

	foundMobius := false
	foundYellow := false
	iterate := func() {
		for itr.next() {
			trackfile := itr.value()
			if trackfile.Contact.UnitID == mobius.Contact.UnitID {
				require.EqualValues(t, mobius, trackfile)
				foundMobius = true
			} else if trackfile.Contact.UnitID == yellow.Contact.UnitID {
				require.EqualValues(t, yellow, trackfile)
				foundYellow = true
			}
		}
	}
	iterate()
	require.True(t, foundMobius)
	require.True(t, foundYellow)

	itr.reset()

	foundMobius = false
	foundYellow = false
	iterate()

}
