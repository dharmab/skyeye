package spatial

import "github.com/dharmab/skyeye/pkg/spatial/projections"

// options holds configuration for spatial calculations.
type options struct {
	projection projections.Projection
}

// Option configures spatial calculations.
type Option func(*options)

// WithProjection uses the given projection for calculations. When a projection
// is provided, distance and bearing calculations use planar geometry on the
// projected coordinate system instead of spherical geometry.
//
// This improves accuracy when working with DCS World coordinates, which use
// Transverse Mercator projections internally.
func WithProjection(p projections.Projection) Option {
	return func(o *options) {
		o.projection = p
	}
}

// applyOptions applies all options to a new options struct.
func applyOptions(opts []Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
