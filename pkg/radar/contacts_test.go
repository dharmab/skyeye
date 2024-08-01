package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
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

	val, ok := d.getByCallsign("mobius 1")
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)

	_, ok = d.getByCallsign("yellow 13")
	require.False(t, ok)
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
