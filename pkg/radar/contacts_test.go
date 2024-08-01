package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/gammazero/deque"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

func TestGetByCallsign(t *testing.T) {
	d := newContactDatabase()
	tf := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf)

	val, ok := d.getByCallsign("mobius 1")
	require.True(t, ok)
	require.EqualValues(t, tf, val)

	_, ok = d.getByCallsign("yellow 13")
	require.False(t, ok)
}

func TestGetByUnitID(t *testing.T) {
	d := newContactDatabase()
	tf := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf)

	val, ok := d.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, tf, val)

	_, ok = d.getByUnitID(2)
	require.False(t, ok)
}

func TestSet(t *testing.T) {
	d := newContactDatabase()
	tf := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf)

	val, ok := d.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, tf, val)

	tf.Update(trackfile.Frame{
		Timestamp: time.Now(),
		Point: orb.Point{
			1,
			1,
		},
		Altitude: unit.Length(1000) * unit.Foot,
		Heading:  unit.Angle(90) * unit.Degree,
	})

	d.set(tf)

	val, ok = d.getByUnitID(1)
	require.True(t, ok)
	require.EqualValues(t, tf, val)
}

func TestDelete(t *testing.T) {
	d := newContactDatabase()
	tf := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf)

	_, ok := d.getByUnitID(1)
	require.True(t, ok)

	d.delete(1)

	_, ok = d.getByUnitID(1)
	require.False(t, ok)
}

func TestItr(t *testing.T) {
	d := newContactDatabase()
	tf1 := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    1,
			Name:      "Mobius 1 Reaper",
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf1)

	tf2 := &trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:    2,
			Name:      "Yellow 13 Reiher",
			Coalition: coalitions.Red,
			ACMIName:  "Su-27",
		},
		Track: *deque.New[trackfile.Frame](),
	}
	d.set(tf2)

	itr := d.itr()

	tf1Found := false
	tf2Found := false
	iterate := func() {
		for itr.next() {
			tf := itr.value()
			if tf.Contact.UnitID == tf1.Contact.UnitID {
				require.EqualValues(t, tf1, tf)
				tf1Found = true
			} else if tf.Contact.UnitID == tf2.Contact.UnitID {
				require.EqualValues(t, tf2, tf)
				tf2Found = true
			}
		}
	}
	iterate()
	require.True(t, tf1Found)
	require.True(t, tf2Found)

	itr.reset()

	tf1Found = false
	tf2Found = false
	iterate()

}
