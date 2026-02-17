package projections

import (
	"math"
	"testing"

	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransverseMercatorRoundTrip(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name   string
		center orb.Point
		point  orb.Point
	}{
		{
			name:   "near center",
			center: orb.Point{-115.0, 36.0}, // Nevada
			point:  orb.Point{-115.5, 36.5},
		},
		{
			name:   "far from center",
			center: orb.Point{-115.0, 36.0},
			point:  orb.Point{-116.0, 37.0},
		},
		{
			name:   "high latitude",
			center: orb.Point{33.0, 69.0}, // Kola
			point:  orb.Point{34.0, 70.0},
		},
		{
			name:   "southern hemisphere",
			center: orb.Point{-59.0, -51.0}, // South Atlantic
			point:  orb.Point{-58.0, -50.0},
		},
		{
			name:   "same as center",
			center: orb.Point{37.0, 45.0},
			point:  orb.Point{37.0, 45.0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			proj := NewTransverseMercator(WithCenter(tc.center))

			// Forward projection
			projected := proj.ToProjected(tc.point)

			// Inverse projection
			recovered := proj.ToWGS84(projected)

			// Should recover original point within tolerance
			// Allow 1e-8 degree tolerance (~1mm at Earth surface)
			assert.InDelta(t, tc.point.Lon(), recovered.Lon(), 1e-8, "longitude mismatch")
			assert.InDelta(t, tc.point.Lat(), recovered.Lat(), 1e-8, "latitude mismatch")
		})
	}
}

func TestTransverseMercatorCenter(t *testing.T) {
	t.Parallel()
	// When projecting the center point, it should be at (0, 0)
	center := orb.Point{37.0, 45.0}
	proj := NewTransverseMercator(WithCenter(center))

	projected := proj.ToProjected(center)

	require.InDelta(t, 0, projected[0], 1e-6, "easting of center should be 0")
	require.InDelta(t, 0, projected[1], 1e-6, "northing of center should be 0")
}

func TestTransverseMercatorDefaults(t *testing.T) {
	t.Parallel()
	// Default projection should have scale factor 1.0
	proj := NewTransverseMercator()

	// Project a point near the origin (0, 0)
	point := orb.Point{0.1, 0.1}
	projected := proj.ToProjected(point)
	recovered := proj.ToWGS84(projected)

	assert.InDelta(t, point.Lon(), recovered.Lon(), 1e-8, "longitude mismatch")
	assert.InDelta(t, point.Lat(), recovered.Lat(), 1e-8, "latitude mismatch")
}

func TestTransverseMercatorWithScaleFactor(t *testing.T) {
	t.Parallel()
	// UTM-style projection with scale factor 0.9996
	proj := NewTransverseMercator(
		WithCentralMeridian(33),
		WithScaleFactor(0.9996),
		WithFalseEasting(-99517),
		WithFalseNorthing(-4998115),
	)

	// Round-trip should still work
	point := orb.Point{34.0, 45.0}
	projected := proj.ToProjected(point)
	recovered := proj.ToWGS84(projected)

	assert.InDelta(t, point.Lon(), recovered.Lon(), 1e-8, "longitude mismatch")
	assert.InDelta(t, point.Lat(), recovered.Lat(), 1e-8, "latitude mismatch")

	// False easting/northing should shift the projected coordinates
	projNoOffset := NewTransverseMercator(
		WithCentralMeridian(33),
		WithScaleFactor(0.9996),
	)
	projectedNoOffset := projNoOffset.ToProjected(point)

	assert.InDelta(t, projectedNoOffset[0]-99517, projected[0], 1e-6, "false easting offset")
	assert.InDelta(t, projectedNoOffset[1]-4998115, projected[1], 1e-6, "false northing offset")
}

func TestTransverseMercatorProjectionInterface(t *testing.T) {
	t.Parallel()
	// Verify TransverseMercator satisfies the Projection interface
	var p Projection = NewTransverseMercator(WithCenter(orb.Point{37.0, 45.0}))
	point := orb.Point{37.1, 45.1}

	projected := p.ToProjected(point)
	recovered := p.ToWGS84(projected)

	assert.InDelta(t, point.Lon(), recovered.Lon(), 1e-8)
	assert.InDelta(t, point.Lat(), recovered.Lat(), 1e-8)
}

func TestTransverseMercatorDistanceConsistency(t *testing.T) {
	t.Parallel()
	// Two projections centered on the same point should produce identical results
	center := orb.Point{37.0, 45.0}
	proj1 := NewTransverseMercator(WithCenter(center))
	proj2 := NewTransverseMercator(
		WithCentralMeridian(center.Lon()),
		WithOriginLatitude(center.Lat()),
	)

	a := orb.Point{37.0, 45.0}
	b := orb.Point{37.1, 45.1}

	projA1 := proj1.ToProjected(a)
	projB1 := proj1.ToProjected(b)
	projA2 := proj2.ToProjected(a)
	projB2 := proj2.ToProjected(b)

	dx1 := projB1[0] - projA1[0]
	dy1 := projB1[1] - projA1[1]
	dist1 := math.Sqrt(dx1*dx1 + dy1*dy1)

	dx2 := projB2[0] - projA2[0]
	dy2 := projB2[1] - projA2[1]
	dist2 := math.Sqrt(dx2*dx2 + dy2*dy2)

	assert.InDelta(t, dist1, dist2, 1e-9, "distances should be identical")
}
