package brevity

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
)

type Vector interface {
	// Bearing is the heading from the fighter to the target location, rounded to the nearest degree.
	Bearing() bearings.Bearing
	// Range is the distance from the fighter to the target location, rounded to the nearest nautical mile.
	Range() unit.Length
}

type vector struct {
	bearing  bearings.Bearing
	distance unit.Length
}

func NewVector(bearing bearings.Bearing, distance unit.Length) Vector {
	return &vector{
		bearing:  bearing,
		distance: distance,
	}
}

// Bearing implements [Vector.Bearing].
func (v *vector) Bearing() bearings.Bearing {
	return v.bearing
}

// Range implements [Vector.Range].
func (v *vector) Range() unit.Length {
	return unit.Length(math.Round(v.distance.NauticalMiles())) * unit.NauticalMile
}

type VectorRequest struct {
	// Callsign of the friendly aircraft requesting the vector.
	Callsign string
	//  Location to which the friendly aircraft is requesting a vector.
	Location string
}

func (r VectorRequest) String() string {
	return "VECTOR to " + r.Location + " for " + r.Callsign
}

type VectorResponse struct {
	// Callsign of the friendly aircraft requesting the vector.
	Callsign string
	// Location which the friendly aircraft is requesting a vector.
	Location string
	// Contact is true if the callsign was correlated to an aircraft on frequency, otherwise false.
	Contact bool
	// Status is true if the vector was successfully computed, otherwise false.
	Status bool
	// Vector is the computed vector to the target location, if available.
	// // If Status is false, this may be nil.
	Vector Vector
}
