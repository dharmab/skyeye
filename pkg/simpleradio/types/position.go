package types

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	Altitude  float64 `json:"alt"`
}
