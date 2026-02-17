// Package projections provides map projections for converting between WGS84
// geographic coordinates and flat projected coordinate systems.
package projections

import "github.com/paulmach/orb"

// Projection converts between WGS84 geographic coordinates and a flat
// projected coordinate system.
type Projection interface {
	// ToProjected converts a WGS84 point (orb.Point{longitude, latitude} in
	// degrees) to projected coordinates (orb.Point{easting, northing} in
	// meters).
	ToProjected(point orb.Point) orb.Point
	// ToWGS84 converts projected coordinates (orb.Point{easting, northing} in
	// meters) back to WGS84 geographic coordinates (orb.Point{longitude,
	// latitude} in degrees).
	ToWGS84(projected orb.Point) orb.Point
}
