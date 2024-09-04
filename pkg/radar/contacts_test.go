package radar

import (
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

func TestGetByCallsign(t *testing.T) {
	t.Parallel()
	db := newContactDatabase()
	trackfile := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        1,
		Name:      "Mobius 1 Reaper",
		Coalition: coalitions.Blue,
		ACMIName:  "F-15C",
	})
	db.set(trackfile)

	name, tf, ok := db.getByCallsignAndCoalititon("mobius 1", coalitions.Blue)
	require.True(t, ok)
	require.Equal(t, "mobius 1", name)
	require.EqualValues(t, trackfile, tf)

	_, _, ok = db.getByCallsignAndCoalititon("mobius 1", coalitions.Red)
	require.False(t, ok)

	name, tf, ok = db.getByCallsignAndCoalititon("moebius 1", coalitions.Blue)
	require.True(t, ok)
	require.Equal(t, "mobius 1", name)
	require.EqualValues(t, trackfile, tf)

	_, _, ok = db.getByCallsignAndCoalititon("yellow 13", coalitions.Red)
	require.False(t, ok)
}

func TestRealCallsigns(t *testing.T) {
	t.Parallel()
	// Callsigns collected from Discord
	testCases := []struct {
		Name    string
		heardAs string
	}{
		{Name: "Hussein 1-1 | SpyderF16", heardAs: "houston 1 1"},
		{Name: "Witch 1-1", heardAs: "which 1 1"},
		{Name: "Spare 15", heardAs: "spear 15"},
		{Name: "Olympus-1-1", heardAs: "olympus 1 1"},
	}
	db := newContactDatabase()

	for i, test := range testCases {
		trackfile := trackfiles.NewTrackfile(trackfiles.Labels{
			ID:        uint64(i),
			Name:      test.Name,
			Coalition: coalitions.Blue,
			ACMIName:  "F-15C",
		})
		db.set(trackfile)
	}

	for i, test := range testCases {
		parsedCallsign, ok := parser.ParsePilotCallsign(test.Name)
		require.True(t, ok)
		foundCallsign, tf, ok := db.getByCallsignAndCoalititon(test.heardAs, coalitions.Blue)
		require.True(t, ok, "queried %s, expected %s, but result was %v", test.heardAs, test.Name, ok)
		require.Equal(t, parsedCallsign, foundCallsign)
		require.EqualValues(t, uint64(i), tf.Contact.ID)
	}
}

func TestGetByID(t *testing.T) {
	t.Parallel()
	db := newContactDatabase()
	trackfile := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        1,
		Name:      "Mobius 1 Reaper",
		Coalition: coalitions.Blue,
		ACMIName:  "F-15C",
	})
	db.set(trackfile)

	val, ok := db.getByID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)

	_, ok = db.getByID(2)
	require.False(t, ok)
}

func TestSet(t *testing.T) {
	t.Parallel()
	database := newContactDatabase()
	trackfile := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        1,
		Name:      "Mobius 1 Reaper",
		Coalition: coalitions.Blue,
		ACMIName:  "F-15C",
	})
	database.set(trackfile)

	val, ok := database.getByID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)

	trackfile.Update(trackfiles.Frame{
		Time: time.Now(),
		Point: orb.Point{
			1,
			1,
		},
		Altitude: 1000 * unit.Foot,
		Heading:  90 * unit.Degree,
	})

	database.set(trackfile)

	val, ok = database.getByID(1)
	require.True(t, ok)
	require.EqualValues(t, trackfile, val)
}

func TestDelete(t *testing.T) {
	t.Parallel()
	database := newContactDatabase()
	trackfile := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        1,
		Name:      "Mobius 1 Reaper",
		Coalition: coalitions.Blue,
		ACMIName:  "F-15C",
	})
	database.set(trackfile)

	_, ok := database.getByID(1)
	require.True(t, ok)

	ok = database.delete(1)
	require.True(t, ok)

	_, ok = database.getByID(1)
	require.False(t, ok)

	ok = database.delete(2)
	require.False(t, ok)
}

func TestValues(t *testing.T) {
	t.Parallel()
	db := newContactDatabase()

	mobius := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        1,
		Name:      "Mobius 1 Reaper",
		Coalition: coalitions.Blue,
		ACMIName:  "F-15C",
	})
	db.set(mobius)

	yellow := trackfiles.NewTrackfile(trackfiles.Labels{
		ID:        2,
		Name:      "Yellow 13 Reiher",
		Coalition: coalitions.Red,
		ACMIName:  "Su-27",
	})
	db.set(yellow)

	foundMobius := false
	foundYellow := false
	for trackfile := range db.values() {
		if trackfile.Contact.ID == mobius.Contact.ID {
			require.EqualValues(t, mobius, trackfile)
			foundMobius = true
		} else if trackfile.Contact.ID == yellow.Contact.ID {
			require.EqualValues(t, yellow, trackfile)
			foundYellow = true
		}
		if foundMobius && foundYellow {
			break
		}
	}
	require.True(t, foundMobius)
	require.True(t, foundYellow)
}
