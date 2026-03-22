package brevity

import (
	"math"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
)

// Vector is a magnetic bearing and distance. It is the base type for BULLSEYE, BRA, and BRAA.
type Vector struct {
	bearing  bearings.Bearing
	distance unit.Length
}

// NewVector creates a new [Vector] from a magnetic bearing and distance.
func NewVector(bearing bearings.Bearing, distance unit.Length) *Vector {
	return &Vector{
		bearing:  bearing,
		distance: distance,
	}
}

// Bearing returns the magnetic bearing, rounded to the nearest degree.
func (v *Vector) Bearing() bearings.Bearing {
	return v.bearing
}

// Range returns the distance, rounded to the nearest nautical mile.
func (v *Vector) Range() unit.Length {
	return unit.Length(math.Round(v.distance.NauticalMiles())) * unit.NauticalMile
}

// VectorRequest is a request for a VECTOR to a named location.
type VectorRequest struct {
	// Callsign of the friendly aircraft requesting the vector.
	Callsign string
	// Location to which the friendly aircraft is requesting a vector.
	Location string
}

func (r VectorRequest) String() string {
	return "VECTOR to " + r.Location + " for " + r.Callsign
}

// VectorResponse is a response to a VECTOR request.
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
	// If Status is false, this may be nil.
	Vector *Vector
}
