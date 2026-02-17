package projections

import (
	"math"

	"github.com/paulmach/orb"
)

// Option configures a TransverseMercator projection.
type Option func(*TransverseMercator)

// WithCentralMeridian sets the longitude (in degrees) of the central meridian.
func WithCentralMeridian(degrees float64) Option {
	return func(tm *TransverseMercator) {
		tm.centralMeridian = degrees * math.Pi / 180.0
	}
}

// WithOriginLatitude sets the latitude (in degrees) used as the northing origin.
func WithOriginLatitude(degrees float64) Option {
	return func(tm *TransverseMercator) {
		tm.originLatitude = degrees * math.Pi / 180.0
	}
}

// WithScaleFactor sets the scale factor at the central meridian. For UTM and
// DCS World projections this is typically 0.9996.
func WithScaleFactor(k float64) Option {
	return func(tm *TransverseMercator) {
		tm.scaleFactor = k
	}
}

// WithFalseEasting sets the constant offset (in meters) added to easting
// coordinates so that all values in the useful area are positive.
func WithFalseEasting(meters float64) Option {
	return func(tm *TransverseMercator) {
		tm.falseEasting = meters
	}
}

// WithFalseNorthing sets the constant offset (in meters) added to northing
// coordinates so that all values in the useful area are positive.
func WithFalseNorthing(meters float64) Option {
	return func(tm *TransverseMercator) {
		tm.falseNorthing = meters
	}
}

// WithCenter sets the central meridian and origin latitude from a geographic
// point (orb.Point{longitude, latitude} in degrees). This is a convenience
// option that sets both values at once.
func WithCenter(center orb.Point) Option {
	return func(tm *TransverseMercator) {
		tm.centralMeridian = center.Lon() * math.Pi / 180.0
		tm.originLatitude = center.Lat() * math.Pi / 180.0
	}
}
