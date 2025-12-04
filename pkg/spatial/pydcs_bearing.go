package spatial

import (
	"fmt"
	"math"

	"github.com/martinlindhe/unit"
	"github.com/michiho/go-proj/v10"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"

	"github.com/dharmab/skyeye/pkg/bearings"
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

// ProjectionToLatLong converts projection coordinates to latitude/longitude using Kola terrain parameters
func ProjectionToLatLong(x, z float64) (float64, float64, error) {
	// Get the Kola projection parameters
	projection := KolaProjection()

	// Create transformer from the Kola projection to WGS84
	// This is the inverse of LatLongToProjection
	source := projection.ToProjString()
	target := "+proj=longlat +datum=WGS84 +no_defs +type=crs"

	pj, err := proj.NewCRSToCRS(source, target, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create projection: %v", err)
	}
	defer pj.Destroy()

	// Create coordinate from x/z (swapped to match the forward transformation)
	// In LatLongToProjection we return (result.Y(), result.X())
	// So here we need to input (z, x) to get back the original (lon, lat)
	coord := proj.NewCoord(z, x, 0, 0)

	// Transform the coordinates (inverse transformation)
	result, err := pj.Forward(coord)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %v", err)
	}

	// Result contains lon, lat (in that order)
	lon := result.X()
	lat := result.Y()

	// Validate output coordinates
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("result latitude out of range: %f", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("result longitude out of range: %f", lon)
	}

	return lat, lon, nil
}

// CalculateDistance calculates the distance between two points in meters
func CalculateDistance(lat1, lon1, lat2, lon2 float64) (float64, error) {
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
	//distanceNauticalMiles := distanceMeters / 1852

	return distanceMeters, nil
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

// PointAtBearingAndDistanceUTM calculates a new point at the given bearing and distance
// from an origin point using Transverse Mercator projection
func PointAtBearingAndDistanceUTM(lat1 float64, lon1 float64, bearing bearings.Bearing, distance unit.Length) orb.Point {
	if bearing.IsMagnetic() {
		log.Warn().Stringer("bearing", bearing).Msg("bearing provided to PointAtBearingAndDistance should not be magnetic")
	}

	// Convert origin to projection coordinates
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		log.Error().Msgf("failed to convert origin point: %v", err)
	}

	// Convert bearing to radians
	bearingRadians := bearing.Degrees() * math.Pi / 180.0

	// Calculate the new position in projection space
	// x is northing (Y from PROJ), z is easting (X from PROJ)
	// For bearing clockwise from North: north = cos(bearing), east = sin(bearing)
	distanceMeters := distance.Meters()
	deltaX := math.Cos(bearingRadians) * distanceMeters
	deltaZ := math.Sin(bearingRadians) * distanceMeters

	x2 := x1 + deltaX
	z2 := z1 + deltaZ

	// Convert back to lat/lon
	lat2, lon2, err := ProjectionToLatLong(x2, z2)
	if err != nil {
		log.Error().Msgf("failed to convert result to lat/lon: %v", err)
	}
	//log.Debug().Float64("lat1", lat1).Float64("lon1", lon1).Msg("message")
	//log.Debug().Float64("lat2", lat2).Float64("lon2", lon2).Msg("message")
	return orb.Point{lon2, lat2}
}

/*
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
*/
