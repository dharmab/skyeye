package types

// Position is a 3D coordinate indicating the source position of a transmission.
type Position struct {
	// Latitude is the north-south coordinate in decimal degrees.
	Latitude float64 `json:"lat"`
	// Longitude is the east-west coordinate in decimal degrees.
	Longitude float64 `json:"lng"`
	// Altitude is the height above sea level in meters.
	Altitude float64 `json:"alt"`
}
