// Package spatial contains geospatial functions.
// It provides functions for working with the [github.com/paulmach/orb],
// [github.com/dharmab/bearings] and [github.com/martinlindhe/unit] modules
// together.
package spatial

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/spatial/projections"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// Distance returns the distance between two points on the Earth's surface.
// By default, this uses great-circle distance on a spherical Earth model.
//
// When working with DCS World coordinates, use WithProjection to match DCS's
// internal coordinate system and improve accuracy at extreme latitudes.
func Distance(a, b orb.Point, opts ...Option) unit.Length {
	o := applyOptions(opts)
	if o.projection != nil {
		return projectedDistance(o.projection, a, b)
	}
	return unit.Length(math.Abs(geo.Distance(a, b))) * unit.Meter
}

// projectedDistance calculates the Euclidean distance between two points
// after projecting them to a Transverse Mercator coordinate system.
func projectedDistance(p projections.Projection, a, b orb.Point) unit.Length {
	projA := p.ToProjected(a)
	projB := p.ToProjected(b)

	dx := projB[0] - projA[0]
	dy := projB[1] - projA[1]

	return unit.Length(math.Sqrt(dx*dx+dy*dy)) * unit.Meter
}

// TrueBearing returns the true bearing from point a to point b.
// By default, this uses geodesic bearing on a spherical Earth model.
//
// When working with DCS World coordinates, use WithProjection to match DCS's
// internal coordinate system and improve accuracy at extreme latitudes.
func TrueBearing(a, b orb.Point, opts ...Option) bearings.Bearing {
	o := applyOptions(opts)
	if o.projection != nil {
		return projectedBearing(o.projection, a, b)
	}
	direction := unit.Angle(geo.Bearing(a, b)) * unit.Degree
	return bearings.NewTrueBearing(direction)
}

// projectedBearing calculates the bearing from point a to point b
// after projecting them to a Transverse Mercator coordinate system.
func projectedBearing(p projections.Projection, a, b orb.Point) bearings.Bearing {
	projA := p.ToProjected(a)
	projB := p.ToProjected(b)

	dx := projB[0] - projA[0]
	dy := projB[1] - projA[1]

	// atan2(dx, dy) gives the angle from north (Y-axis) clockwise
	angle := math.Atan2(dx, dy) * 180.0 / math.Pi

	return bearings.NewTrueBearing(unit.Angle(angle) * unit.Degree)
}

// PointAtBearingAndDistance returns the point at the given bearing and distance
// from the origin.
// By default, this uses spherical Earth geometry.
//
// When working with DCS World coordinates, use WithProjection to match DCS's
// internal coordinate system and improve accuracy at extreme latitudes.
func PointAtBearingAndDistance(origin orb.Point, bearing bearings.Bearing, distance unit.Length, opts ...Option) orb.Point {
	if bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to PointAtBearingAndDistance should not be magnetic")
	}
	o := applyOptions(opts)
	if o.projection != nil {
		return projectedPointAtBearingAndDistance(o.projection, origin, bearing, distance)
	}
	return geo.PointAtBearingAndDistance(origin, bearing.Degrees(), distance.Meters())
}

// projectedPointAtBearingAndDistance calculates a point at the given bearing
// and distance from the origin using planar geometry on a Transverse Mercator
// projection.
func projectedPointAtBearingAndDistance(p projections.Projection, origin orb.Point, bearing bearings.Bearing, distance unit.Length) orb.Point {
	// Project origin to TM coordinates
	projOrigin := p.ToProjected(origin)

	// Convert bearing to radians (from north, clockwise)
	bearingRad := bearing.Degrees() * math.Pi / 180.0

	// Calculate offset in meters
	dx := distance.Meters() * math.Sin(bearingRad)
	dy := distance.Meters() * math.Cos(bearingRad)

	// New point in projected coordinates
	projPoint := orb.Point{
		projOrigin[0] + dx,
		projOrigin[1] + dy,
	}

	// Convert back to WGS84
	return p.ToWGS84(projPoint)
}

// IsZero returns true if the point is the origin.
func IsZero(point orb.Point) bool {
	return point.Equal(orb.Point{})
}

// NormalizeAltitude returns the absolute length rounded to the nearest 1000 feet, or nearest 100 feet if less than 1000 feet.
func NormalizeAltitude(altitude unit.Length) unit.Length {
	if altitude < 0 {
		altitude = -altitude
	}
	bucketWidth := 1000 * unit.Foot
	if altitude < bucketWidth {
		bucketWidth = 100. * unit.Foot
	}
	bucket := int(math.Round(altitude.Feet() / bucketWidth.Feet()))
	rounded := int(bucketWidth.Feet()) * bucket
	return unit.Length(rounded) * unit.Foot
}
