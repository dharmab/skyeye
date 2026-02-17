// Package terrains contains data for DCS World terrain maps.
package terrains

import (
	"math"

	"github.com/dharmab/skyeye/pkg/spatial/projections"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
)

// Terrain name constants for DCS World maps.
const (
	Caucasus       = "Caucasus"
	Nevada         = "Nevada"
	PersianGulf    = "PersianGulf"
	Normandy       = "Normandy"
	TheChannel     = "TheChannel"
	Syria          = "Syria"
	MarianaIslands = "MarianaIslands"
	SouthAtlantic  = "SouthAtlantic"
	Sinai          = "Sinai"
	Kola           = "Kola"
	Afghanistan    = "Afghanistan"
	Iraq           = "Iraq"
	ColdWarGermany = "GermanyCW"
)

// All contains all known DCS terrains with their Transverse Mercator projection parameters.
// Parameters are sourced from pydcs (https://github.com/pydcs/dcs) and
// dcs-web-editor-mono (https://github.com/DCS-Web-Editor/dcs-web-editor-mono).
var All = []Terrain{
	{Name: Caucasus, CentralMeridian: 33, FalseEasting: -99517, FalseNorthing: -4998115},
	{Name: Nevada, CentralMeridian: -117, FalseEasting: -193996.81, FalseNorthing: -4410028.064},
	{Name: PersianGulf, CentralMeridian: 57, FalseEasting: 75756, FalseNorthing: -2894933},
	{Name: Normandy, CentralMeridian: -3, FalseEasting: -195526, FalseNorthing: -5484813},
	{Name: TheChannel, CentralMeridian: 3, FalseEasting: 99376, FalseNorthing: -5636889},
	{Name: Syria, CentralMeridian: 39, FalseEasting: 282801, FalseNorthing: -3879866},
	{Name: MarianaIslands, CentralMeridian: 147, FalseEasting: 238418, FalseNorthing: -1491840},
	{Name: SouthAtlantic, CentralMeridian: -57, FalseEasting: 147640, FalseNorthing: 5815417},
	{Name: Sinai, CentralMeridian: 33, FalseEasting: 169222, FalseNorthing: -3325313},
	{Name: Kola, CentralMeridian: 21, FalseEasting: -62702, FalseNorthing: -7543625},
	{Name: Afghanistan, CentralMeridian: 63, FalseEasting: -300150, FalseNorthing: -3759657},
	{Name: Iraq, CentralMeridian: 45, FalseEasting: 72290, FalseNorthing: -3680057},
	{Name: ColdWarGermany, CentralMeridian: 21, FalseEasting: 35427.62, FalseNorthing: -6061633.13},
}

// scaleFactor is the scale factor at the central meridian (k_0) used by all DCS terrains.
// This is the standard UTM scale factor.
const scaleFactor = 0.9996

// metersPerDegreeLatitude is an approximation of the average distance in meters per degree of latitude.
// This is used as part of the heuristic for guessing which terrain is in use when Tacview doesn't export the terrain directly.
const metersPerDegreeLatitude = 111320.0

// Terrain represents a DCS World terrain/map with its Transverse Mercator projection parameters.
type Terrain struct {
	// Name is the terrain identifier.
	Name string
	// CentralMeridian is the longitude of the projection's central meridian in degrees.
	CentralMeridian float64
	// FalseEasting is added to all X coordinates in meters.
	FalseEasting float64
	// FalseNorthing is added to all Y coordinates in meters.
	FalseNorthing float64
}

// Center returns the approximate geographic center of the terrain.
// This is derived from the central meridian and false northing to provide
// a representative point for the terrain. Useful for terrain selection.
func (t Terrain) Center() orb.Point {
	// The false northing gives us an approximate latitude via meridional arc.
	// This approximation is sufficient for selecting the closest terrain.
	approxLat := -t.FalseNorthing / metersPerDegreeLatitude
	return orb.Point{t.CentralMeridian, approxLat}
}

// Projection returns a Transverse Mercator projection configured for this terrain.
func (t Terrain) Projection() *projections.TransverseMercator {
	return projections.NewTransverseMercator(
		projections.WithCentralMeridian(t.CentralMeridian),
		projections.WithScaleFactor(scaleFactor),
		projections.WithFalseEasting(t.FalseEasting),
		projections.WithFalseNorthing(t.FalseNorthing),
	)
}

// Closest returns the terrain whose center is closest to the given point.
// This is used to determine which Transverse Mercator projection to use
// when the terrain is not explicitly known (e.g., when Tacview doesn't export MapId).
func Closest(point orb.Point) Terrain {
	var closest Terrain
	minDistance := math.MaxFloat64

	for _, terrain := range All {
		distance := geo.Distance(point, terrain.Center())
		if distance < minDistance {
			minDistance = distance
			closest = terrain
		}
	}

	return closest
}
