package main

import (
	"fmt"
	"math"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

func main() {
	// In-game coordinates
	playerLat := 69.047461
	playerLon := 33.405794
	targetLat := 69.157219
	targetLon := 32.14515

	// Points in orb.Point format (lon, lat)
	playerPoint := orb.Point{playerLon, playerLat}
	targetPoint := orb.Point{targetLon, targetLat}

	fmt.Printf("Player Point: Lat=%f, Lon=%f\n", playerPoint.Lat(), playerPoint.Lon())
	fmt.Printf("Target Point: Lat=%f, Lon=%f\n", targetPoint.Lat(), targetPoint.Lon())

	// Calculate distance using great circle
	greatCircleDistance := geo.Distance(playerPoint, targetPoint)
	fmt.Printf("Distance (great circle): %f meters (%f nautical miles)\n", greatCircleDistance, greatCircleDistance*0.000539957)

	// Calculate bearing using great circle
	greatCircleBearing := geo.Bearing(playerPoint, targetPoint)
	// Normalize bearing to 0-360 degrees
	if greatCircleBearing < 0 {
		greatCircleBearing += 360
	}
	fmt.Printf("Bearing (great circle): %f degrees\n", greatCircleBearing)

	// Test with reversed coordinates (lat, lon instead of lon, lat)
	playerPointReversed := orb.Point{playerLat, playerLon}
	targetPointReversed := orb.Point{targetLat, targetLon}

	fmt.Printf("\nReversed coordinates:\n")
	fmt.Printf("Player Point: Lat=%f, Lon=%f\n", playerPointReversed.Lat(), playerPointReversed.Lon())
	fmt.Printf("Target Point: Lat=%f, Lon=%f\n", targetPointReversed.Lat(), targetPointReversed.Lon())

	// Calculate distance using great circle with reversed coordinates
	greatCircleDistanceReversed := geo.Distance(playerPointReversed, targetPointReversed)
	fmt.Printf("Distance (great circle, reversed): %f meters (%f nautical miles)\n", greatCircleDistanceReversed, greatCircleDistanceReversed*0.000539957)

	// Calculate bearing using great circle with reversed coordinates
	greatCircleBearingReversed := geo.Bearing(playerPointReversed, targetPointReversed)
	// Normalize bearing to 0-360 degrees
	if greatCircleBearingReversed < 0 {
		greatCircleBearingReversed += 360
	}
	fmt.Printf("Bearing (great circle, reversed): %f degrees\n", greatCircleBearingReversed)

	// Expected values from in-game
	expectedBearing := 273.0
	expectedDistanceNM := 188.0

	fmt.Printf("\nExpected Bearing: %f degrees\n", expectedBearing)
	fmt.Printf("Expected Distance: %f nautical miles\n", expectedDistanceNM)

	// Calculate differences for normal coordinates
	bearingDiff := math.Abs(greatCircleBearing - expectedBearing)
	distanceDiffNM := math.Abs(greatCircleDistance*0.000539957 - expectedDistanceNM)

	fmt.Printf("\nNormal Coordinates - Bearing Difference: %f degrees\n", bearingDiff)
	fmt.Printf("Normal Coordinates - Distance Difference: %f nautical miles\n", distanceDiffNM)

	// Calculate differences for reversed coordinates
	bearingDiffReversed := math.Abs(greatCircleBearingReversed - expectedBearing)
	distanceDiffNMReversed := math.Abs(greatCircleDistanceReversed*0.000539957 - expectedDistanceNM)

	fmt.Printf("Reversed Coordinates - Bearing Difference: %f degrees\n", bearingDiffReversed)
	fmt.Printf("Reversed Coordinates - Distance Difference: %f nautical miles\n", distanceDiffNMReversed)
}
