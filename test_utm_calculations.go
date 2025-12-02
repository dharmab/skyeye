package main

import (
	"fmt"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/im7mortal/UTM"
	"github.com/paulmach/orb"
)

func main() {
	// Test data from the requirements
	// Player aircraft coordinates: N 69°02'50.86"   E 33°24'20.86"
	// 69.047461 33.405794
	playerLat := 69.047461
	playerLon := 33.405794
	playerPoint := orb.Point{playerLon, playerLat}

	// Bullseye coordinates: N 68°28'27.91"   E 22°52'01.66"
	// bullseye declination +6.8
	bullseyeLat := 68.474419 // 68°28'27.91"
	bullseyeLon := 22.867128 // 22°52'01.66"
	bullseyePoint := orb.Point{bullseyeLon, bullseyeLat}

	// Target coordinates: N 69°09'25.99"   E 32°08'42.54"
	// 69.157219 32.14515
	targetLat := 69.545253
	targetLon := 24.858169
	targetPoint := orb.Point{targetLon, targetLat}

	// Same-grid target:
	// Lat Long Precise: N 64°55'07.14"   E 34°15'46.76"
	// 64.91865 34.262989
	sameGridTargetLat := 64.91865
	sameGridTargetLon := 34.262989
	sameGridTargetPoint := orb.Point{sameGridTargetLon, sameGridTargetLat}

	fmt.Println("=== UTM Conversion Test ===")
	testUTMConversion(playerPoint, "Player")
	testUTMConversion(bullseyePoint, "Bullseye")
	testUTMConversion(targetPoint, "Target")
	testUTMConversion(sameGridTargetPoint, "Same-grid target")
	fmt.Println("\n=== Distance and Bearing Calculations ===")

	// Test player to target (different UTM zones)
	fmt.Println("\n--- Player to Target (Different UTM zones) ---")
	distance := spatial.Distance(playerPoint, targetPoint)
	bearing := spatial.TrueBearing(playerPoint, targetPoint)
	fmt.Printf("Distance: %.2f nautical miles\n", distance.NauticalMiles())
	fmt.Printf("Bearing: %.2f degrees true\n", bearing.Degrees())
	fmt.Printf("Expected: ~188 nautical miles, ~273 degrees true\n")

	// Test player to same-grid target (same UTM zone)
	fmt.Println("\n--- Player to Same-grid Target (Same UTM zone) ---")
	distance2 := spatial.Distance(playerPoint, sameGridTargetPoint)
	bearing2 := spatial.TrueBearing(playerPoint, sameGridTargetPoint)
	fmt.Printf("Distance: %.2f nautical miles\n", distance2.NauticalMiles())
	fmt.Printf("Bearing: %.2f degrees true\n", bearing2.Degrees())

	// Test declination values
	fmt.Println("\n=== Declination Values ===")
	testDeclination(playerPoint, 12.8, "Player")
	testDeclination(bullseyePoint, 6.8, "Bullseye")
	testDeclination(targetPoint, 12.1, "Target")
	testDeclination(sameGridTargetPoint, 11.5, "Same-grid target")
}

func testUTMConversion(point orb.Point, name string) {
	easting, northing, zoneNumber, zoneLetter, err := UTM.FromLatLon(point.Lat(), point.Lon(), point.Lat() >= 0)
	if err != nil {
		fmt.Printf("%s: Error converting to UTM: %v\n", name, err)
		return
	}
	fmt.Printf("%s: Zone %d%s, Easting: %.2f, Northing: %.2f\n", name, zoneNumber, zoneLetter, easting, northing)
}

func testDeclination(point orb.Point, expectedDeclination float64, name string) {
	// Using a fixed date for consistent results
	t := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	declination, err := bearings.Declination(point, t)
	if err != nil {
		fmt.Printf("%s: Error getting declination: %v\n", name, err)
		return
	}
	fmt.Printf("%s: Declination %.1f° (expected %.1f°)\n", name, declination.Degrees(), expectedDeclination)
}
