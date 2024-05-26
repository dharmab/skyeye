package encyclopedia

import "github.com/paulmach/orb"

// Terrain describes mapping data for a DCS terrain.
// This data is used to project between DCS's flat earth coordinate system and Long/Lat.
// Values are chosen to match the DCS Web Editor for ease of creating test data.
type Terrain struct {
	// Friendly name of the terrain.
	Name string
	// Name of the terrain as it appears in mission Lua files.
	EditorName string
	// ProjectionBounds is the edge of the terrain when projected into the transverse Mercator projection.
	ProjectionBounds orb.Ring
	// CentralMeridian is the central meridian of the terrain when projected into the transverse Mercator projection.
	CentralMeridian float64
	// FalseEasting is the longitude of the terrain's origin point.
	FalseEasting float64
	// FalseNorthing is the latitude of the terrain's origin point.
	FalseNorthing float64
}

// Values from https://github.com/DCS-Web-Editor/dcs-web-editor-mono/blob/main/packages/map-projection/src/index.ts

var Caucases = Terrain{
	Name:       "Caucasus",
	EditorName: "Caucasus",
	ProjectionBounds: orb.Ring{
		orb.Point{26.778743595881, 48.387663480938},
		orb.Point{39.608931903399, 27.637331401126},
		orb.Point{47.142314272867, 38.86511140611},
		orb.Point{49.309787386754, 47.382221906262},
		orb.Point{26.778743595881, 48.387663480938},
	},
	CentralMeridian: 33,
	FalseEasting:    -99516.9999999732,
	FalseNorthing:   -4998114.999999984,
}
