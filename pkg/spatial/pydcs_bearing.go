package spatial

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/martinlindhe/unit"
	"github.com/michiho/go-proj/v10"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"

	"github.com/dharmab/skyeye/pkg/bearings"
)

func greatCircleDeg(lat1, lon1, lat2, lon2 float64) float64 {
	lat1r := lat1 * math.Pi / 180
	lon1r := lon1 * math.Pi / 180
	lat2r := lat2 * math.Pi / 180
	lon2r := lon2 * math.Pi / 180
	dLat := lat2r - lat1r
	dLon := lon2r - lon1r
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1r)*math.Cos(lat2r)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return c // radians
}

// TransverseMercator represents the parameters for a Transverse Mercator projection.
type TransverseMercator struct {
	CentralMeridian int
	FalseEasting    float64
	FalseNorthing   float64
	ScaleFactor     float64
}

type latLonBounds struct {
	minLat float64
	maxLat float64
	minLon float64
	maxLon float64
}

type terrainDef struct {
	name      string
	tm        TransverseMercator
	boundsXY  [4]float64 // x1, y1, x2, y2 in projected coordinates (DCS x/y)
	latLonBox latLonBounds
	centerLat float64
	centerLon float64
}

func (l latLonBounds) contains(lat, lon float64) bool {
	return lat >= l.minLat && lat <= l.maxLat && lon >= l.minLon && lon <= l.maxLon
}

func (l latLonBounds) area() float64 {
	return math.Abs(l.maxLat-l.minLat) * math.Abs(l.maxLon-l.minLon)
}

var (
	projectionMu      sync.RWMutex
	currentProjection = CaucasusProjection()
	currentTerrain    = "Caucasus"
	terrainDetected   atomic.Bool
	bullseyes         = make(map[string]orb.Point) // coalition or source label -> bullseye
)

var terrainDefs = []terrainDef{
	{name: "Afghanistan", tm: AfghanistanProjection(), boundsXY: [4]float64{532000.0, -534000.0, -512000.0, 757000.0}, centerLat: 33.9346, centerLon: 66.24705},
	{name: "Caucasus", tm: CaucasusProjection(), boundsXY: [4]float64{380 * 1000, -560 * 1000, -600 * 1000, 1130 * 1000}, centerLat: 43.69666, centerLon: 32.96},
	{name: "Falklands", tm: FalklandsProjection(), boundsXY: [4]float64{74967, -114995, -129982, 129991}, centerLat: 52.468, centerLon: 59.173},
	{name: "GermanyCW", tm: GermanyColdWarProjection(), boundsXY: [4]float64{260000.0, -1100000.0, -600000.0, -425000.0}, centerLat: 51.0, centerLon: 11.0},
	{name: "Iraq", tm: IraqProjection(), boundsXY: [4]float64{440000.0, -500000.0, -950000.0, 850000.0}, centerLat: 30.76, centerLon: 59.07},
	{name: "Kola", tm: KolaProjection(), boundsXY: [4]float64{-315000, -890000, 900000, 856000}, centerLat: 68.0, centerLon: 22.5},
	{name: "MarianaIslands", tm: MarianasProjection(), boundsXY: [4]float64{1000 * 10000, -1000 * 1000, -300 * 1000, 500 * 1000}, centerLat: 13.485, centerLon: 144.798},
	{name: "Nevada", tm: NevadaProjection(), boundsXY: [4]float64{-167000.0, -330000.0, -500000.0, 210000.0}, centerLat: 39.81806, centerLon: -114.73333},
	{name: "Normandy", tm: NormandyProjection(), boundsXY: [4]float64{-132707.843750, -389942.906250, 185756.156250, 165065.078125}, centerLat: 41.3, centerLon: 0.18},
	{name: "PersianGulf", tm: PersianGulfProjection(), boundsXY: [4]float64{-218768.750000, -392081.937500, 197357.906250, 333129.125000}, centerLat: 0, centerLon: 0},
	{name: "Sinai", tm: SinaiProjection(), boundsXY: [4]float64{-450000, -280000, 500000, 560000}, centerLat: 30.047, centerLon: 31.224},
	{name: "Syria", tm: SyriaProjection(), boundsXY: [4]float64{-320000, -579986, 300000, 579998}, centerLat: 35.021, centerLon: 35.901},
	{name: "TheChannel", tm: TheChannelProjection(), boundsXY: [4]float64{74967, -114995, -129982, 129991}, centerLat: 50.875, centerLon: 1.5875},
}

func init() {
	for i := range terrainDefs {
		if err := computeLatLonBounds(&terrainDefs[i]); err != nil {
			log.Warn().Err(err).Str("terrain", terrainDefs[i].name).Msg("failed to compute lat/lon bounds for terrain")
		}
	}
}

func terrainDefByName(name string) (terrainDef, bool) {
	for _, td := range terrainDefs {
		if td.name == name {
			return td, true
		}
	}
	return terrainDef{}, false
}

func computeLatLonBounds(td *terrainDef) error {
	// boundsXY are DCS projected coords: x=easting, y=northing in meters.
	x1, y1, x2, y2 := td.boundsXY[0], td.boundsXY[1], td.boundsXY[2], td.boundsXY[3]
	norths := []float64{y1, y2}
	easts := []float64{x1, x2}

	minLat := math.Inf(1)
	maxLat := math.Inf(-1)
	minLon := math.Inf(1)
	maxLon := math.Inf(-1)

	for _, north := range norths {
		for _, east := range easts {
			lat, lon, err := ProjectionToLatLongFor(td.tm, north, east)
			if err != nil {
				return fmt.Errorf("convert bounds corner: %w", err)
			}
			if lat < minLat {
				minLat = lat
			}
			if lat > maxLat {
				maxLat = lat
			}
			if lon < minLon {
				minLon = lon
			}
			if lon > maxLon {
				maxLon = lon
			}
		}
	}

	td.latLonBox = latLonBounds{
		minLat: minLat,
		maxLat: maxLat,
		minLon: minLon,
		maxLon: maxLon,
	}
	if td.centerLat == 0 && td.centerLon == 0 {
		td.centerLat = (minLat + maxLat) / 2
		td.centerLon = (minLon + maxLon) / 2
	}
	return nil
}

func bullseyeInsideBounds(td terrainDef, bullseye orb.Point) bool {
	if td.latLonBox.contains(bullseye.Lat(), bullseye.Lon()) {
		return true
	}

	xMin := math.Min(td.boundsXY[0], td.boundsXY[2])
	xMax := math.Max(td.boundsXY[0], td.boundsXY[2])
	yMin := math.Min(td.boundsXY[1], td.boundsXY[3])
	yMax := math.Max(td.boundsXY[1], td.boundsXY[3])

	x, z, err := LatLongToProjectionFor(td.tm, bullseye.Lat(), bullseye.Lon())
	if err == nil {
		north := x
		east := z
		if east >= xMin && east <= xMax && north >= yMin && north <= yMax {
			return true
		}
	}

	return false
}

func setCurrentTerrain(name string, tm TransverseMercator) {
	projectionMu.Lock()
	defer projectionMu.Unlock()
	if currentTerrain != name {
		log.Debug().
			Str("from", currentTerrain).
			Str("to", name).
			Msg("switching terrain projection")
	}
	currentTerrain = name
	currentProjection = tm
}

// ForceTerrain overrides the current terrain selection and disables auto-detection.
func ForceTerrain(name string, tm TransverseMercator) {
	setCurrentTerrain(name, tm)
	terrainDetected.Store(true)
	projectionMu.Lock()
	bullseyes = make(map[string]orb.Point)
	projectionMu.Unlock()
}

// ResetTerrainToDefault resets terrain selection to the default (Caucasus) and re-enables auto-detection.
func ResetTerrainToDefault() {
	setCurrentTerrain("Caucasus", CaucasusProjection())
	terrainDetected.Store(false)
	projectionMu.Lock()
	bullseyes = make(map[string]orb.Point)
	projectionMu.Unlock()
}

func getCurrentProjection() TransverseMercator {
	projectionMu.RLock()
	defer projectionMu.RUnlock()
	return currentProjection
}

func allBullseyesInside(td terrainDef, points []orb.Point) bool {
	for _, p := range points {
		if !bullseyeInsideBounds(td, p) {
			return false
		}
	}
	return true
}

// DetectTerrainFromBullseye attempts to pick the terrain based on all known bullseyes.
// Provide a source label (e.g., coalition) to track multiple bullseyes. Returns whether the terrain changed.
func DetectTerrainFromBullseye(source string, bullseye orb.Point) (string, bool) {
	projectionMu.Lock()
	bullseyes[source] = bullseye
	current := currentTerrain
	points := make([]orb.Point, 0, len(bullseyes))
	for _, p := range bullseyes {
		points = append(points, p)
	}
	projectionMu.Unlock()

	// If current terrain fits all bullseyes, no change.
	detected := terrainDetected.Load()
	if detected {
		if td, ok := terrainDefByName(current); ok && allBullseyesInside(td, points) {
			return current, false
		}
	}

	// Pick the smallest-area terrain that contains all bullseyes.
	bestName := ""
	bestTM := TransverseMercator{}
	bestArea := math.Inf(1)

	for _, td := range terrainDefs {
		if !allBullseyesInside(td, points) {
			continue
		}
		area := td.latLonBox.area()
		if area == 0 || math.IsNaN(area) || math.IsInf(area, 0) {
			area = math.Abs(td.boundsXY[0]-td.boundsXY[2]) * math.Abs(td.boundsXY[1]-td.boundsXY[3])
		}
		if area < bestArea || (area == bestArea && td.name < bestName) {
			bestArea = area
			bestName = td.name
			bestTM = td.tm
		}
	}

	if bestName != "" {
		changed := !detected || bestName != current
		setCurrentTerrain(bestName, bestTM)
		terrainDetected.Store(true)
		return bestName, changed
	}

	// Fallback: pick the terrain with minimal total center distance to all bullseyes.
	minTotal := math.Inf(1)
	for _, td := range terrainDefs {
		total := 0.0
		for _, p := range points {
			total += greatCircleDeg(p.Lat(), p.Lon(), td.centerLat, td.centerLon)
		}
		if total < minTotal || (total == minTotal && td.name < bestName) {
			minTotal = total
			bestName = td.name
			bestTM = td.tm
		}
	}

	if bestName != "" {
		changed := !detected || bestName != current
		setCurrentTerrain(bestName, bestTM)
		terrainDetected.Store(true)
		return bestName, changed
	}

	// No terrain fits or can be inferred.
	return "", false
}

// Terrain projection parameter helpers (sourced from pydcs terrain definitions).
func AfghanistanProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 63,
		FalseEasting:    -300149.9999999864,
		FalseNorthing:   -3759657.000000049,
		ScaleFactor:     0.9996,
	}
}

func CaucasusProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 33,
		FalseEasting:    -99516.9999999732,
		FalseNorthing:   -4998114.999999984,
		ScaleFactor:     0.9996,
	}
}

func FalklandsProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: -57,
		FalseEasting:    147639.99999997593,
		FalseNorthing:   5815417.000000032,
		ScaleFactor:     0.9996,
	}
}

func GermanyColdWarProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 21,
		FalseEasting:    35427.619999985734,
		FalseNorthing:   -6061633.128000011,
		ScaleFactor:     0.9996,
	}
}

func IraqProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 45,
		FalseEasting:    72290.00000004497,
		FalseNorthing:   -3680057.0,
		ScaleFactor:     0.9996,
	}
}

// KolaProjection returns the TransverseMercator parameters for the Kola terrain.
func KolaProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 21,
		FalseEasting:    -62702.00000000087,
		FalseNorthing:   -7543624.999999979,
		ScaleFactor:     0.9996,
	}
}

func MarianasProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 147,
		FalseEasting:    238417.99999989968,
		FalseNorthing:   -1491840.000000048,
		ScaleFactor:     0.9996,
	}
}

func NevadaProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: -117,
		FalseEasting:    -193996.80999964548,
		FalseNorthing:   -4410028.063999966,
		ScaleFactor:     0.9996,
	}
}

func NormandyProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: -3,
		FalseEasting:    -195526.00000000204,
		FalseNorthing:   -5484812.999999951,
		ScaleFactor:     0.9996,
	}
}

func PersianGulfProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 57,
		FalseEasting:    75755.99999999645,
		FalseNorthing:   -2894933.0000000377,
		ScaleFactor:     0.9996,
	}
}

func SinaiProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 33,
		FalseEasting:    169221.9999999585,
		FalseNorthing:   -3325312.9999999693,
		ScaleFactor:     0.9996,
	}
}

func SyriaProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 39,
		FalseEasting:    282801.00000003993,
		FalseNorthing:   -3879865.9999999935,
		ScaleFactor:     0.9996,
	}
}

func TheChannelProjection() TransverseMercator {
	return TransverseMercator{
		CentralMeridian: 3,
		FalseEasting:    99376.00000000288,
		FalseNorthing:   -5636889.00000001,
		ScaleFactor:     0.9996,
	}
}

// ToProjString converts the TransverseMercator parameters to a PROJ string.
func (tm TransverseMercator) ToProjString() string {
	return fmt.Sprintf(
		"+proj=tmerc +lat_0=0 +lon_0=%d +k=%f +x_0=%f +y_0=%f +ellps=WGS84 +towgs84=0,0,0,0,0,0,0 +units=m +no_defs +type=crs",
		tm.CentralMeridian,
		tm.ScaleFactor,
		tm.FalseEasting,
		tm.FalseNorthing,
	)
}

// LatLongToProjection converts latitude/longitude to projection coordinates using the current terrain parameters.
func LatLongToProjection(lat float64, lon float64) (x float64, z float64, err error) {
	return LatLongToProjectionFor(getCurrentProjection(), lat, lon)
}

// LatLongToProjectionFor converts latitude/longitude to projection coordinates using the provided projection parameters.
func LatLongToProjectionFor(tm TransverseMercator, lat float64, lon float64) (x float64, z float64, err error) {
	// Validate input coordinates
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180, got %f", lon)
	}

	// Create transformer from WGS84 to the projection.
	// Using the exact PROJ string from the Python implementation.
	source := "+proj=longlat +datum=WGS84 +no_defs +type=crs"
	target := tm.ToProjString()

	pj, err := proj.NewCRSToCRS(source, target, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create projection: %w", err)
	}
	defer pj.Destroy()

	// Create coordinate from lon/lat (PROJ uses lon,lat order).
	coord := proj.NewCoord(lon, lat, 0, 0)

	// Transform the coordinates
	result, err := pj.Forward(coord)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %w", err)
	}

	// In DCS, z coordinate corresponds to the y coordinate from projection.
	// But in our case, we need to swap x and y to match the Python results.
	return result.Y(), result.X(), nil
}

// ProjectionToLatLong converts projection coordinates to latitude/longitude using the current terrain parameters.
func ProjectionToLatLong(x, z float64) (lat float64, lon float64, err error) {
	return ProjectionToLatLongFor(getCurrentProjection(), x, z)
}

// ProjectionToLatLongFor converts projection coordinates to latitude/longitude using the provided projection parameters.
func ProjectionToLatLongFor(tm TransverseMercator, x, z float64) (lat float64, lon float64, err error) {
	// Create transformer from the projection to WGS84.
	// This is the inverse of LatLongToProjection.
	source := tm.ToProjString()
	target := "+proj=longlat +datum=WGS84 +no_defs +type=crs"

	pj, err := proj.NewCRSToCRS(source, target, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create projection: %w", err)
	}
	defer pj.Destroy()

	// Create coordinate from x/z (swapped to match the forward transformation).
	// In LatLongToProjection we return (result.Y(), result.X()).
	// So here we need to input (z, x) to get back the original (lon, lat).
	coord := proj.NewCoord(z, x, 0, 0)

	// Transform the coordinates (inverse transformation).
	result, err := pj.Forward(coord)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to transform coordinates: %w", err)
	}

	// Result contains lon, lat (in that order).
	lon = result.X()
	lat = result.Y()

	// Validate output coordinates.
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("result latitude out of range: %f", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("result longitude out of range: %f", lon)
	}

	return lat, lon, nil
}

// CalculateDistance calculates the distance between two points in meters.
func CalculateDistance(lat1, lon1, lat2, lon2 float64) (float64, error) {
	// Convert both points to projection coordinates.
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %w", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %w", err)
	}

	// Calculate Euclidean distance in meters.
	dx := x2 - x1
	dz := z2 - z1
	distanceMeters := math.Sqrt(dx*dx + dz*dz)

	// Convert meters to nautical miles (1 nautical mile = 1852 meters).
	//distanceNauticalMiles := distanceMeters / 1852.

	return distanceMeters, nil
}

// CalculateBearing calculates the true bearing from first point to second point using projection coordinates.
func CalculateBearing(lat1, lon1, lat2, lon2 float64) (float64, error) {
	// Convert both points to projection coordinates.
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %w", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %w", err)
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

// PointAtBearingAndDistanceUTM calculates a new point at the given bearing and distance.
// from an origin point using Transverse Mercator projection.
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
