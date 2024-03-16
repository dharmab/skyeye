package brevity

import "github.com/martinlindhe/unit"

// PictureRequest is a request for an updated PICTURE.
type PictureRequest struct {
	// Callsign of the friendly aircraft requesting the PICTURE.
	Callsign string
	// Radius is the distance from the friendly aircraft to search for groups.
	// This is present to allow server admins to cap the scale of a PICTURE call, since some DCS servers are quite dense.
	Radius unit.Length
}

// PICTURE is a report to establish a tactical air image.
// Reference: ATP 3-52.4 Chapter IV section 9
type PictureResponse struct {
	// Groups included in the PICTURE.
	Groups []Group
}
