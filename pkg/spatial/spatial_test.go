// spatial contains functions for working with the orb, bearings and unit modules together.
package spatial

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed spatial_test.json
var spatialTestJSON []byte

type coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (c coordinate) point() orb.Point {
	return orb.Point{c.Lon, c.Lat}
}

type bearingSet struct {
	AB     float64 `json:"ab"`
	AC     float64 `json:"ac"`
	CB     float64 `json:"cb"`
	BullsA float64 `json:"bullsA"`
	BullsB float64 `json:"bullsB"`
	BullsC float64 `json:"bullsC"`
}

type terrainFixture struct {
	Terrain string `json:"terrain"`
	Points  struct {
		A        coordinate `json:"a"`
		B        coordinate `json:"b"`
		C        coordinate `json:"c"`
		Bullseye coordinate `json:"bullseye"`
	} `json:"points"`
	Bearings struct {
		True     bearingSet `json:"true"`
		Magnetic bearingSet `json:"magnetic"`
	} `json:"bearings"`
	Distances struct {
		AB     float64 `json:"ab"`
		AC     float64 `json:"ac"`
		CB     float64 `json:"cb"`
		BullsA float64 `json:"bullsA"`
		BullsB float64 `json:"bullsB"`
		BullsC float64 `json:"bullsC"`
	} `json:"distances"`
}

type spatialTestFixtures struct {
	TestData struct {
		Date     string           `json:"date"`
		Terrains []terrainFixture `json:"terrains"`
	} `json:"test data"`
}

func loadSpatialFixtures(t *testing.T) []terrainFixture {
	t.Helper()

	var fixtures spatialTestFixtures
	err := json.Unmarshal(spatialTestJSON, &fixtures)
	require.NoError(t, err, "failed to decode spatial_test.json")
	require.NotEmpty(t, fixtures.TestData.Terrains, "no terrain fixtures loaded")

	return fixtures.TestData.Terrains
}

func terrainByName(name string) (terrainDef, bool) {
	for _, td := range terrainDefs {
		if strings.EqualFold(td.name, name) {
			return td, true
		}
	}
	return terrainDef{}, false
}

func TestMain(m *testing.M) {
	ForceTerrain("Kola", KolaProjection())
	code := m.Run()
	ResetTerrainToDefault()
	os.Exit(code)
}

//nolint:paralleltest // serialized because tests mutate global terrain state
func TestDistance(t *testing.T) {
	testCases := loadSpatialFixtures(t)

	for _, terrain := range testCases {
		t.Run(terrain.Terrain, func(t *testing.T) {
			td, ok := terrainByName(terrain.Terrain)
			require.True(t, ok, "unknown terrain %s", terrain.Terrain)
			ForceTerrain(td.name, td.tm)
			t.Cleanup(func() {
				ForceTerrain("Kola", KolaProjection())
			})

			bullseye := terrain.Points.Bullseye.point()
			cases := []struct {
				name     string
				a        orb.Point
				b        orb.Point
				expected unit.Length
			}{
				{
					name:     "ab",
					a:        terrain.Points.A.point(),
					b:        terrain.Points.B.point(),
					expected: unit.Length(terrain.Distances.AB) * unit.NauticalMile,
				},
				{
					name:     "ac",
					a:        terrain.Points.A.point(),
					b:        terrain.Points.C.point(),
					expected: unit.Length(terrain.Distances.AC) * unit.NauticalMile,
				},
				{
					name:     "bc",
					a:        terrain.Points.C.point(),
					b:        terrain.Points.B.point(),
					expected: unit.Length(terrain.Distances.CB) * unit.NauticalMile,
				},
				{
					name:     "bullsA",
					a:        bullseye,
					b:        terrain.Points.A.point(),
					expected: unit.Length(terrain.Distances.BullsA) * unit.NauticalMile,
				},
				{
					name:     "bullsB",
					a:        bullseye,
					b:        terrain.Points.B.point(),
					expected: unit.Length(terrain.Distances.BullsB) * unit.NauticalMile,
				},
				{
					name:     "bullsC",
					a:        bullseye,
					b:        terrain.Points.C.point(),
					expected: unit.Length(terrain.Distances.BullsC) * unit.NauticalMile,
				},
			}

			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					actual := Distance(test.a, test.b)
					assert.InDelta(t, test.expected.NauticalMiles(), actual.NauticalMiles(), 5)
				})
			}
		})
	}
}

//nolint:paralleltest // serial because tests mutate global terrain state
func TestTrueBearing(t *testing.T) {
	testCases := loadSpatialFixtures(t)

	for _, terrain := range testCases {
		t.Run(terrain.Terrain, func(t *testing.T) {
			td, ok := terrainByName(terrain.Terrain)
			require.True(t, ok, "unknown terrain %s", terrain.Terrain)
			ForceTerrain(td.name, td.tm)
			t.Cleanup(func() {
				ForceTerrain("Kola", KolaProjection())
			})

			bullseye := terrain.Points.Bullseye.point()
			cases := []struct {
				name     string
				a        orb.Point
				b        orb.Point
				expected unit.Angle
			}{
				{
					name:     "ab",
					a:        terrain.Points.A.point(),
					b:        terrain.Points.B.point(),
					expected: unit.Angle(terrain.Bearings.True.AB) * unit.Degree,
				},
				{
					name:     "ac",
					a:        terrain.Points.A.point(),
					b:        terrain.Points.C.point(),
					expected: unit.Angle(terrain.Bearings.True.AC) * unit.Degree,
				},
				{
					name:     "bc",
					a:        terrain.Points.C.point(),
					b:        terrain.Points.B.point(),
					expected: unit.Angle(terrain.Bearings.True.CB) * unit.Degree,
				},
				{
					name:     "bullsA",
					a:        bullseye,
					b:        terrain.Points.A.point(),
					expected: unit.Angle(terrain.Bearings.True.BullsA) * unit.Degree,
				},
				{
					name:     "bullsB",
					a:        bullseye,
					b:        terrain.Points.B.point(),
					expected: unit.Angle(terrain.Bearings.True.BullsB) * unit.Degree,
				},
				{
					name:     "bullsC",
					a:        bullseye,
					b:        terrain.Points.C.point(),
					expected: unit.Angle(terrain.Bearings.True.BullsC) * unit.Degree,
				},
			}

			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					actual := TrueBearing(test.a, test.b)
					assert.InDelta(t, test.expected.Degrees(), actual.Degrees(), 2)
				})
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		p        orb.Point
		expected bool
	}{
		{
			p:        orb.Point{0, 0},
			expected: true,
		},
		{
			p:        orb.Point{0, 1},
			expected: false,
		},
		{
			p:        orb.Point{1, 0},
			expected: false,
		},
		{
			p:        orb.Point{1, 1},
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v", test.p), func(t *testing.T) {
			t.Parallel()
			actual := IsZero(test.p)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestPointAtBearingAndDistance(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		origin   orb.Point
		bearing  bearings.Bearing
		distance unit.Length
		expected orb.Point
	}{
		{
			origin:   orb.Point{0, 0},
			bearing:  bearings.NewTrueBearing(90 * unit.Degree),
			distance: 0,
			expected: orb.Point{0, 0},
		},
		{
			origin:   orb.Point{0, 0},
			bearing:  bearings.NewTrueBearing(90 * unit.Degree),
			distance: 111 * unit.Kilometer,
			expected: orb.Point{1, 0},
		},
		{
			origin:   orb.Point{22.867128, 68.474419},
			bearing:  bearings.NewTrueBearing(75 * unit.Degree),
			distance: 430 * unit.Kilometer,
			expected: orb.Point{33.405794, 69.047461},
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%v, %v, %v", test.origin, test.bearing, test.distance), func(t *testing.T) {
			t.Parallel()
			actual := PointAtBearingAndDistance(test.origin, test.bearing, test.distance)
			assert.InDelta(t, test.expected.Lon(), actual.Lon(), 1.5)
			assert.InDelta(t, test.expected.Lat(), actual.Lat(), 1.5)
		})
	}
}

func TestNormalizeAltitude(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input    unit.Length
		expected unit.Length
	}{
		{
			input:    -100 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    0,
			expected: 0,
		},
		{
			input:    40 * unit.Foot,
			expected: 0,
		},
		{
			input:    100 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    120 * unit.Foot,
			expected: 100 * unit.Foot,
		},
		{
			input:    200 * unit.Foot,
			expected: 200 * unit.Foot,
		},
		{
			input:    249 * unit.Foot,
			expected: 200 * unit.Foot,
		},
		{
			input:    250 * unit.Foot,
			expected: 300 * unit.Foot,
		},
		{
			input:    1234 * unit.Foot,
			expected: 1000 * unit.Foot,
		},
		{
			input:    10000 * unit.Foot,
			expected: 10000 * unit.Foot,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("%fft", test.input.Feet()), func(t *testing.T) {
			t.Parallel()
			actual := NormalizeAltitude(test.input)
			assert.InDelta(t, test.expected.Feet(), actual.Feet(), 0.1)
		})
	}
}

func TestProjectionRoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		lat float64
		lon float64
	}{
		{lat: 69.047461, lon: 33.405794},
		{lat: 70.068836, lon: 24.973478},
		{lat: 64.91865, lon: 34.262989},
		{lat: 68.474419, lon: 22.867128},
		{lat: 0, lon: 0},
		{lat: 45, lon: 45},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("lat=%f,lon=%f", test.lat, test.lon), func(t *testing.T) {
			t.Parallel()
			x, z, err := LatLongToProjection(test.lat, test.lon)
			require.NoError(t, err)

			lat2, lon2, err := ProjectionToLatLong(x, z)
			require.NoError(t, err)

			assert.InDelta(t, test.lat, lat2, 0.000001, "latitude mismatch")
			assert.InDelta(t, test.lon, lon2, 0.000001, "longitude mismatch")
		})
	}
}
