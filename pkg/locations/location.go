package locations

import "github.com/paulmach/orb"

type Location struct {
	Names     []string `json:"names"`
	Longitude float64  `json:"longitude"`
	Latitude  float64  `json:"latitude"`
}

func (l Location) Point() orb.Point {
	return orb.Point{l.Longitude, l.Latitude}
}
