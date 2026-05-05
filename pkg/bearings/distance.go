package bearings

import (
	"math"

	"github.com/martinlindhe/unit"
)

// AngularDistance returns the smallest angular difference between two bearings, accounting for
// wrap-around (e.g. 001° and 359° are 2° apart, not 358°). The returned angle is always in the
// range [0°, 180°]. Both bearings are compared by their normalized value; the caller is
// responsible for ensuring they are expressed in the same reference frame (both true or both
// magnetic), since no declination conversion is performed.
func AngularDistance(a, b Bearing) unit.Angle {
	diff := math.Abs(a.Value().Degrees() - b.Value().Degrees())
	if diff > 180 {
		diff = 360 - diff
	}
	return unit.Angle(diff) * unit.Degree
}
