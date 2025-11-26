package main

import (
	"fmt"
	"time"

	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/proway2/go-igrf/igrf"
)

func main() {
	// Your aircraft position: N 69°02'44" E 33°24'14"
	// Converting to decimal degrees:
	// 69°02'44" = 69 + 2/60 + 44/3600 = 69.045555...°
	// 33°24'14" = 33 + 24/60 + 14/3600 = 33.403888...°
	// Note: orb.Point is [longitude, latitude]
	origin := orb.Point{33.40388888888889, 69.04555555555555} // lon, lat

	// Target aircraft position: N 69°33'47" E 27°36'23"
	// 69°33'47" = 69 + 33/60 + 47/3600 = 69.563055...°
	// 27°36'23" = 27 + 36/60 + 23/3600 = 27.606388...°
	// Note: orb.Point is [longitude, latitude]
	target := orb.Point{27.60638888888889, 69.56305555555555} // lon, lat

	fmt.Printf("Origin (your aircraft): %.8f°N, %.8f°E\n", origin.Lat(), origin.Lon())
	fmt.Printf("Target (enemy aircraft): %.8f°N, %.8f°E\n", target.Lat(), target.Lon())

	// Date: 1999-06-11
	t := time.Date(1999, 6, 11, 0, 0, 0, 0, time.UTC)
	fmt.Printf("Date: %s\n", t.Format("2006-01-02"))

	// Calculate true bearing
	trueBearing := spatial.TrueBearing(origin, target)
	fmt.Printf("True bearing: %.1f°\n", trueBearing.Degrees())

	// Calculate declination at origin (your aircraft position)
	igrd := igrf.New()
	// Using decimal year for 1999-06-11 (day 162 of 1999)
	decimalYear := 1999.0 + 162.0/365.0
	fmt.Printf("Decimal year: %.4f\n", decimalYear)

	field, err := igrd.IGRF(origin.Lat(), origin.Lon(), 0, decimalYear)
	if err != nil {
		fmt.Printf("Error calculating declination: %v\n", err)
		return
	}
	declination := unit.Angle(field.Declination) * unit.Degree
	fmt.Printf("Declination at origin: %.1f°\n", declination.Degrees())

	// Calculate magnetic bearing
	magneticBearing := trueBearing.Magnetic(declination)
	fmt.Printf("Magnetic bearing: %.1f°\n", magneticBearing.Degrees())

	fmt.Printf("\nExpected results:\n")
	fmt.Printf("  Magnetic bearing: 266°\n")
	fmt.Printf("  True bearing: 275°\n")
	fmt.Printf("  Distance: 129nm\n")

	// Calculate distance
	distance := spatial.Distance(origin, target)
	fmt.Printf("\nCalculated distance: %.0f nm\n", distance.NauticalMiles())

	// Let's also test with your stated values to see what would be needed
	fmt.Printf("\nTesting with your stated values:\n")
	fmt.Printf("If true bearing is 275° and declination is 12.8°:\n")
	fmt.Printf("  Magnetic bearing would be: %.1f°\n", 275.0-12.8)
	fmt.Printf("If magnetic bearing is 266° and declination is 12.8°:\n")
	fmt.Printf("  True bearing would be: %.1f°\n", 266.0+12.8)
}
