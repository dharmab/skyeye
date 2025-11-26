package main

import (
	"fmt"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/spatial"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/proway2/go-igrf/igrf"
)

func main() {
	// Your aircraft position: N 69°02'44" E 33°24'14"
	// Target aircraft position: N 69°33'47" E 27°36'23"
	// Note: orb.Point is [longitude, latitude]
	origin := orb.Point{33.40388888888889, 69.04555555555555} // lon, lat
	target := orb.Point{27.60638888888889, 69.56305555555555} // lon, lat
	
	// Date: 1999-06-11
	t := time.Date(1999, 6, 11, 0, 0, 0, 0, time.UTC)
	
	fmt.Printf("=== Coordinate Analysis ===\n")
	fmt.Printf("Origin (your aircraft): %.8f°N, %.8f°E\n", origin.Lat(), origin.Lon())
	fmt.Printf("Target (enemy aircraft): %.8f°N, %.8f°E\n", target.Lat(), target.Lon())
	fmt.Printf("Date: %s\n", t.Format("2006-01-02"))
	
	// Step 1: Calculate true bearing (what SkyEye does)
	trueBearing := spatial.TrueBearing(origin, target)
	fmt.Printf("\n=== Bearing Calculation ===\n")
	fmt.Printf("True bearing (from origin to target): %.1f°\n", trueBearing.Degrees())
	
	// Step 2: Calculate declination at origin (what SkyEye does)
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
	
	// Step 3: Convert to magnetic bearing (what SkyEye does)
	magneticBearing := trueBearing.Magnetic(declination)
	fmt.Printf("Magnetic bearing (true bearing - declination): %.1f°\n", magneticBearing.Degrees())
	
	// Step 4: Verify the conversion
	fmt.Printf("\n=== Verification ===\n")
	fmt.Printf("Verification: %.1f° (true) - %.1f° (declination) = %.1f° (magnetic)\n", 
		trueBearing.Degrees(), declination.Degrees(), trueBearing.Degrees()-declination.Degrees())
	
	// Step 5: Compare with expected values
	fmt.Printf("\n=== Comparison with Expected Values ===\n")
	fmt.Printf("SkyEye result: 274°\n")
	fmt.Printf("Our calculation: %.1f°\n", magneticBearing.Degrees())
	fmt.Printf("Expected result: 266°\n")
	
	// Step 6: What if we use your stated values?
	fmt.Printf("\n=== Using Your Stated Values ===\n")
	yourDeclination := unit.Angle(12.8) * unit.Degree
	yourMagneticBearing := trueBearing.Magnetic(yourDeclination)
	fmt.Printf("Using your stated declination (12.8°): %.1f°\n", yourMagneticBearing.Degrees())
	
	// Step 7: What if the bearing calculation is wrong?
	fmt.Printf("\n=== What If True Bearing Was 275°? ===\n")
	expectedTrueBearing := bearings.NewTrueBearing(275 * unit.Degree)
	expectedMagneticBearing := expectedTrueBearing.Magnetic(declination)
	fmt.Printf("If true bearing was 275°: magnetic = %.1f°\n", expectedMagneticBearing.Degrees())
	
	expectedMagneticBearing2 := expectedTrueBearing.Magnetic(yourDeclination)
	fmt.Printf("If true bearing was 275° and declination 12.8°: magnetic = %.1f°\n", expectedMagneticBearing2.Degrees())
	
	// Step 8: Distance calculation
	fmt.Printf("\n=== Distance Calculation ===\n")
	distance := spatial.Distance(origin, target)
	fmt.Printf("Distance: %.0f nm\n", distance.NauticalMiles())
	fmt.Printf("Expected distance: 129 nm\n")
}