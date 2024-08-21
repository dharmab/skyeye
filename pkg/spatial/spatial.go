// spatial contains functions for working with the orb, bearings and unit modules together.
package spatial

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/rs/zerolog/log"
)

// Distance returns the absolute distance between two points on the earth.
func Distance(a, b orb.Point) unit.Length {
	return unit.Length(math.Abs(geo.Distance(a, b))) * unit.Meter
}

// TrueBearing returns the true bearing between two points.
func TrueBearing(a, b orb.Point) bearings.Bearing {
	return bearings.NewTrueBearing(
		unit.Angle(
			geo.Bearing(a, b),
		) * unit.Degree,
	)
}

func PointAtBearingAndDistance(origin orb.Point, bearing bearings.Bearing, distance unit.Length) orb.Point {
	if bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to PointAtBearingAndDistance should not be magnetic")
	}
	degrees := bearing.Degrees()
	meters := distance.Meters()
	return geo.PointAtBearingAndDistance(origin, degrees, meters)
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
