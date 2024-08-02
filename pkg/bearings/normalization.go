package bearings

import (
	"fmt"
	"math"

	"github.com/martinlindhe/unit"
)

// Normalize returns the normalized angle in the range (0, 360] degrees.
func Normalize(a unit.Angle) unit.Angle {
	θ := a.Degrees()
	for θ < 0 {
		θ += 360
	}
	θ = math.Mod(θ, 360)
	if θ == 0 {
		θ = 360
	}
	return unit.Angle(θ) * unit.Degree
}

func toString(b Bearing) string {
	return fmt.Sprintf("%03.0f", b.RoundedDegrees())
}
