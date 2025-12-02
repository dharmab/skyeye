package main

import (
	"fmt"
	"math"

	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/paulmach/orb"
)

func main() {
	// Test data from the request
	// A coordinates: N 69°02'50.86"   E 33°24'20.86"
	// 69.047461 33.405794
	pointA := orb.Point{33.405794, 69.047461}

	// B coordinates: Lat Long Precise: N 70°04'07.81"   E 24°58'24.52"
	// 70.068836 24.973478
	pointB := orb.Point{24.973478, 70.068836}

	// C coordinates: Lat Long Precise: N 64°55'07.14"   E 34°15'46.76"
	// 64.91865 34.262989
	pointC := orb.Point{34.262989, 64.91865}

	fmt.Println("Testing Distance and Bearing calculations:")
	fmt.Println("=========================================")

	// Test A -> B
	testDistanceAndBearing("A -> B", pointA, pointB, 186, 282)

	// Test A -> C
	testDistanceAndBearing("A -> C", pointA, pointC, 249, 164)

	// Test C -> B
	testDistanceAndBearing("C -> B", pointC, pointB, 377, 317)
}

func testDistanceAndBearing(name string, from, to orb.Point, expectedDistance, expectedBearing int) {
	distance := spatial.Distance(from, to)
	bearing := spatial.TrueBearing(from, to)

	distanceNM := distance.NauticalMiles()
	bearingDegrees := bearing.Degrees()

	fmt.Printf("%s:\n", name)
	fmt.Printf("  Distance: %.0f nautical miles (expected: %d)\n", distanceNM, expectedDistance)
	fmt.Printf("  Bearing: %.0f degrees true (expected: %d)\n", bearingDegrees, expectedBearing)

	// Check if results are within acceptable range
	distanceDiff := math.Abs(distanceNM - float64(expectedDistance))
	bearingDiff := math.Abs(bearingDegrees - float64(expectedBearing))

	if distanceDiff <= 5 && bearingDiff <= 5 {
		fmt.Printf("  Result: PASS (within tolerance)\n")
	} else {
		fmt.Printf("  Result: FAIL (outside tolerance)\n")
	}
	fmt.Println()
}
