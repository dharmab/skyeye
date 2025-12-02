package spatial

import (
	"fmt"
	"math"

	"github.com/michiho/go-proj/v10"
)

// TransverseMercator represents the parameters for a Transverse Mercator projection
type TransverseMercator struct {
	CentralMeridian int
	FalseEasting    float64
	FalseNorthing   float64
	ScaleFactor     float64
}

// KolaProjection returns the TransverseMercator parameters for the Kola terrain
func KolaProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 21,
		FalseEasting:    -62702.00000000087,
		FalseNorthing:   -7543624.999999979,
		ScaleFactor:     0.9996,
	}
}

// ToProjString converts the TransverseMercator parameters to a PROJ string
func (tm TransverseMercator) ToProjString() string {
	return fmt.Sprintf(
		"+proj=tmerc +lat_0=0 +lon_0=%d +k=%f +x_0=%f +y_0=%f +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs",
		tm.CentralMeridian,
		tm.ScaleFactor,
		tm.FalseEasting,
		tm.FalseNorthing,
	)
}

// LatLongToProjection converts latitude/longitude to projection coordinates using Kola terrain parameters
func LatLongToProjection(lat, lon float64) (float64, float64, error) {
	// Validate input coordinates
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180, got %f", lon)
	}

	// Get the Kola projection parameters
	projection := KolaProjection()

	// Create transformer from WGS84 to the Kola projection
	// Using the exact PROJ string from the Python implementation
	source := "+proj=longlat +datum=WGS84 +no_defs +type=crs"
	target := projection.ToProjString()

	pj, err := proj.NewCRSToCRS(source, target, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create projection: %v", err)
	}
	defer pj.Destroy()

	// Create coordinate from lon/lat (PROJ uses lon,lat order)
	coord := proj.NewCoord(lon, lat, 0, 0)

	// Transform the coordinates
	result, err := pj.Forward(coord)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %v", err)
	}

	// In DCS, z coordinate corresponds to the y coordinate from projection
	// But in our case, we need to swap x and y to match the Python results
	return result.Y(), result.X(), nil
}

// CalculateDistanceNauticalMiles calculates the distance between two points in nautical miles
func CalculateDistanceNauticalMiles(lat1, lon1, lat2, lon2 float64) (float64, error) {
	// Convert both points to projection coordinates
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %v", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %v", err)
	}

	// Calculate Euclidean distance in meters
	distanceMeters := math.Sqrt(math.Pow(x2-x1, 2) + math.Pow(z2-z1, 2))

	// Convert meters to nautical miles (1 nautical mile = 1852 meters)
	distanceNauticalMiles := distanceMeters / 1852

	return distanceNauticalMiles, nil
}

// CalculateBearing calculates the true bearing from first point to second point using projection coordinates
func CalculateBearing(lat1, lon1, lat2, lon2 float64) (float64, error) {
	// Convert both points to projection coordinates
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %v", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %v", err)
	}

	// Calculate bearing using atan2
	deltaX := x2 - x1
	deltaZ := z2 - z1

	// atan2 returns angle in radians, convert to degrees
	bearingRadians := math.Atan2(deltaX, deltaZ)
	bearingDegrees := bearingRadians * 180 / math.Pi

	// Convert to compass bearing (0° = North, 90° = East)
	compassBearing := math.Mod(90-bearingDegrees, 360)

	// Ensure bearing is positive
	if compassBearing < 0 {
		compassBearing += 360
	}

	return compassBearing, nil
}

func main() {
	fmt.Println("Distance Calculator using Kola Terrain Projection")
	fmt.Println("==================================================")

	// Example points (Kola map coordinates)
	testCases := []struct {
		lat1, lon1, lat2, lon2 float64
		description            string
	}{
		{69.047461, 33.405794, 70.068836, 24.973478, "A -> B"},
		{69.047461, 33.405794, 64.91865, 34.262989, "A -> C"},
		{64.91865, 34.262989, 70.068836, 24.973478, "C -> B"},
		{65.0, 20.0, 65.0, 20.0, "Same point (zero distance)"},
	}

	for _, tc := range testCases {
		distance, err := CalculateDistanceNauticalMiles(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
		if err != nil {
			fmt.Printf("Error calculating distance for %s: %v\n", tc.description, err)
			continue
		}

		bearing, err := CalculateBearing(tc.lat1, tc.lon1, tc.lat2, tc.lon2)
		if err != nil {
			fmt.Printf("Error calculating bearing for %s: %v\n", tc.description, err)
			continue
		}

		fmt.Printf("%s:\n", tc.description)
		fmt.Printf("  (%f, %f) to (%f, %f)\n", tc.lat1, tc.lon1, tc.lat2, tc.lon2)
		fmt.Printf("  Distance: %.2f nautical miles\n", distance)
		fmt.Printf("  Bearing: %.1f°\n", bearing)
		fmt.Println()
	}
}
