package spatial

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"

	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/project"
	"github.com/rs/zerolog/log"

	"github.com/dharmab/skyeye/pkg/bearings"
)

const (
	wgs84A  = 6378137.0
	wgs84F  = 1 / 298.257223563
	wgs84E2 = wgs84F * (2 - wgs84F)
)

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

type tmProjector struct {
	tm TransverseMercator
}

func (p tmProjector) forward(latDeg, lonDeg float64) (easting, northing float64) {
	lat := deg2rad(latDeg)
	lon := deg2rad(lonDeg)
	origin := deg2rad(float64(p.tm.CentralMeridian))
	k0 := p.tm.ScaleFactor
	coeffs := tmCoefficients()

	lam := lon - origin
	tau := math.Tan(lat)
	taup := taupFromTau(tau)

	sinLam := math.Sin(lam)
	cosLam := math.Cos(lam)
	denom := math.Sqrt(taup*taup + cosLam*cosLam)

	xiPrime := math.Atan2(taup, cosLam)
	etaPrime := math.Asinh(sinLam / denom)

	xi := xiPrime
	eta := etaPrime
	for j := 1; j <= 6; j++ {
		twoJXi := 2 * float64(j) * xiPrime
		twoJEta := 2 * float64(j) * etaPrime
		xi += coeffs.alpha[j] * math.Sin(twoJXi) * math.Cosh(twoJEta)
		eta += coeffs.alpha[j] * math.Cos(twoJXi) * math.Sinh(twoJEta)
	}

	easting = p.tm.FalseEasting + k0*coeffs.a1*eta
	northing = p.tm.FalseNorthing + k0*coeffs.a1*xi
	return easting, northing
}

func (p tmProjector) inverse(easting, northing float64) (latDeg, lonDeg float64) {
	k0 := p.tm.ScaleFactor
	coeffs := tmCoefficients()

	x := (easting - p.tm.FalseEasting) / (k0 * coeffs.a1)
	y := (northing - p.tm.FalseNorthing) / (k0 * coeffs.a1)

	xiPrime := y
	etaPrime := x
	for j := 1; j <= 6; j++ {
		twoJXi := 2 * float64(j) * y
		twoJEta := 2 * float64(j) * x
		xiPrime -= coeffs.beta[j] * math.Sin(twoJXi) * math.Cosh(twoJEta)
		etaPrime -= coeffs.beta[j] * math.Cos(twoJXi) * math.Sinh(twoJEta)
	}

	sinXi := math.Sin(xiPrime)
	cosXi := math.Cos(xiPrime)
	sinhEta := math.Sinh(etaPrime)

	taup := sinXi / math.Sqrt(sinhEta*sinhEta+cosXi*cosXi)
	lam := math.Atan2(sinhEta, cosXi)
	tau := tauFromTaup(taup)

	lat := math.Atan(tau)
	lon := deg2rad(float64(p.tm.CentralMeridian)) + lam

	latDeg = rad2deg(lat)
	lonDeg = rad2deg(lon)
	return latDeg, lonDeg
}

type tmSeriesCoefficients struct {
	a1    float64
	alpha [7]float64
	beta  [7]float64
}

func tmCoefficients() tmSeriesCoefficients {
	n := wgs84F / (2 - wgs84F)
	n2 := n * n
	n3 := n2 * n
	n4 := n2 * n2
	n5 := n4 * n
	n6 := n3 * n3

	a1 := wgs84A / (1 + n) * (1 + n2/4 + n4/64 + n6/256)

	var alpha [7]float64
	alpha[1] = n/2 - 2*n2/3 + 5*n3/16 + 41*n4/180 - 127*n5/288 + 7891*n6/37800
	alpha[2] = 13*n2/48 - 3*n3/5 + 557*n4/1440 + 281*n5/630 - 1983433*n6/1935360
	alpha[3] = 61*n3/240 - 103*n4/140 + 15061*n5/26880 + 167603*n6/181440
	alpha[4] = 49561*n4/161280 - 179*n5/168 + 6601661*n6/7257600
	alpha[5] = 34729*n5/80640 - 3418889*n6/1995840
	alpha[6] = 212378941 * n6 / 319334400

	var beta [7]float64
	beta[1] = n/2 - 2*n2/3 + 37*n3/96 - n4/360 - 81*n5/512 + 96199*n6/604800
	beta[2] = n2/48 + n3/15 - 437*n4/1440 + 46*n5/105 - 1118711*n6/3870720
	beta[3] = 17*n3/480 - 37*n4/840 - 209*n5/4480 + 5569*n6/90720
	beta[4] = 4397*n4/161280 - 11*n5/504 - 830251*n6/7257600
	beta[5] = 4583*n5/161280 - 108847*n6/3991680
	beta[6] = 20648693 * n6 / 638668800

	return tmSeriesCoefficients{
		a1:    a1,
		alpha: alpha,
		beta:  beta,
	}
}

func taupFromTau(tau float64) float64 {
	if tau == 0 {
		return 0
	}
	sinPhi := tau / math.Sqrt(1+tau*tau)
	return math.Sinh(math.Asinh(tau) - math.Sqrt(wgs84E2)*math.Atanh(math.Sqrt(wgs84E2)*sinPhi))
}

func tauFromTaup(taup float64) float64 {
	if taup == 0 {
		return 0
	}
	tau := taup
	for range 10 {
		sqrt1pTau2 := math.Sqrt(1 + tau*tau)
		sinPhi := tau / sqrt1pTau2
		e := math.Sqrt(wgs84E2)
		taupCalc := math.Sinh(math.Asinh(tau) - e*math.Atanh(e*sinPhi))
		delta := taupCalc - taup
		if math.Abs(delta) < 1e-14 {
			break
		}
		denom := 1 - wgs84E2*sinPhi*sinPhi
		duDtau := 1/sqrt1pTau2 - (wgs84E2/(sqrt1pTau2*sqrt1pTau2*sqrt1pTau2))/denom
		dtaupDtau := math.Sqrt(1+taupCalc*taupCalc) * duDtau
		tau -= delta / dtaupDtau
	}
	return tau
}

func (p tmProjector) toProjection() orb.Projection {
	return func(point orb.Point) orb.Point {
		easting, northing := p.forward(point.Lat(), point.Lon())
		return orb.Point{easting, northing}
	}
}

func (p tmProjector) toWGS84() orb.Projection {
	return func(point orb.Point) orb.Point {
		lat, lon := p.inverse(point[0], point[1])
		return orb.Point{lon, lat}
	}
}

func greatCircleDeg(lat1, lon1, lat2, lon2 float64) float64 {
	lat1r := lat1 * math.Pi / 180
	lon1r := lon1 * math.Pi / 180
	lat2r := lat2 * math.Pi / 180
	lon2r := lon2 * math.Pi / 180
	dLat := lat2r - lat1r
	dLon := lon2r - lon1r
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1r)*math.Cos(lat2r)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return c
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
	boundsXY  [4]float64
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
	bullseyes         = make(map[string]orb.Point)
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

	detected := terrainDetected.Load()
	if detected {
		if td, ok := terrainDefByName(current); ok && allBullseyesInside(td, points) {
			return current, false
		}
	}

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

// KolaProjection returns the Transverse Mercator parameters for the Kola terrain.
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

// LatLongToProjection converts latitude/longitude to projection coordinates using the current terrain parameters.
func LatLongToProjection(lat float64, lon float64) (x float64, z float64, err error) {
	return LatLongToProjectionFor(getCurrentProjection(), lat, lon)
}

// LatLongToProjectionFor converts latitude/longitude to projection coordinates using the provided projection parameters.
func LatLongToProjectionFor(tm TransverseMercator, lat float64, lon float64) (x float64, z float64, err error) {
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90, got %f", lat)
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180, got %f", lon)
	}

	tmProj := tmProjector{tm: tm}.toProjection()
	projected := project.Point(orb.Point{lon, lat}, tmProj)
	return projected[1], projected[0], nil
}

// ProjectionToLatLong converts projection coordinates to latitude/longitude using the current terrain parameters.
func ProjectionToLatLong(x, z float64) (lat float64, lon float64, err error) {
	return ProjectionToLatLongFor(getCurrentProjection(), x, z)
}

// ProjectionToLatLongFor converts projection coordinates to latitude/longitude using the provided projection parameters.
func ProjectionToLatLongFor(tm TransverseMercator, x, z float64) (lat float64, lon float64, err error) {
	tmProj := tmProjector{tm: tm}.toWGS84()
	geographic := project.Point(orb.Point{z, x}, tmProj)
	lon = geographic.Lon()
	lat = geographic.Lat()

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
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %w", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %w", err)
	}

	dx := x2 - x1
	dz := z2 - z1
	distanceMeters := math.Sqrt(dx*dx + dz*dz)

	return distanceMeters, nil
}

// CalculateBearing calculates the true bearing from first point to second point using projection coordinates.
func CalculateBearing(lat1, lon1, lat2, lon2 float64) (float64, error) {
	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		return 0, fmt.Errorf("failed to convert first point: %w", err)
	}

	x2, z2, err := LatLongToProjection(lat2, lon2)
	if err != nil {
		return 0, fmt.Errorf("failed to convert second point: %w", err)
	}

	deltaX := x2 - x1
	deltaZ := z2 - z1

	bearingRadians := math.Atan2(deltaX, deltaZ)
	bearingDegrees := bearingRadians * 180 / math.Pi

	compassBearing := math.Mod(90-bearingDegrees, 360)
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

	x1, z1, err := LatLongToProjection(lat1, lon1)
	if err != nil {
		log.Error().Msgf("failed to convert origin point: %v", err)
	}

	bearingRadians := bearing.Degrees() * math.Pi / 180.0

	distanceMeters := distance.Meters()
	deltaX := math.Cos(bearingRadians) * distanceMeters
	deltaZ := math.Sin(bearingRadians) * distanceMeters

	x2 := x1 + deltaX
	z2 := z1 + deltaZ

	lat2, lon2, err := ProjectionToLatLong(x2, z2)
	if err != nil {
		log.Error().Msgf("failed to convert result to lat/lon: %v", err)
	}
	return orb.Point{lon2, lat2}
}
