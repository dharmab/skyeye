package brevity

// PictureRequest is a request for an updated PICTURE.
type PictureRequest struct {
	// Callsign of the friendly aircraft requesting the PICTURE.
	Callsign string
}

func (r PictureRequest) String() string {
	if r.Callsign == "" {
		return "PICTURE"
	}
	return "PICTURE for " + r.Callsign
}

// PICTURE is a report to establish a tactical air image.
// Reference: ATP 3-52.4 Chapter IV section 9.
type PictureResponse struct {
	// Count is the total number of groups in the PICTURE.
	Count int
	// Groups included in the PICTURE. This is a maximum of 3 groups.
	Groups []Group
}
