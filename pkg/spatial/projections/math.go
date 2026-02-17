package projections

import "math"

func square(x float64) float64 {
	return math.Pow(x, 2)
}

func cube(x float64) float64 {
	return math.Pow(x, 3)
}
