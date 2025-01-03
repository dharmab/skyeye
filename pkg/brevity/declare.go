package brevity

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
)

// Declaration identifies a contact's friendly status.
// Reference: ATP 3-52.4 Chapter V section 6.
type Declaration string

const (
	// Bogey indicates the contact's' whose identity is unknown.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Bogey Declaration = "bogey"
	// Friendly indicates the contact is a positively identified friendly.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Friendly Declaration = "friendly"
	// Neutral indicates the contact is a positively identified aircraft whose characteristics, behavior, origin or nationality indicate it is neither supporting nor opposing friendly forces.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Neutral Declaration = "neutral"
	// Bandit indicates the contact is a positively idenfieid enemy in accordance with theater identification criteria. It does not imply direction or authority to engage.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Bandit Declaration = "bandit"
	// Hostile indicates the contact is a identified as an enemy upon which clearance to fire is authorized in accordance with theater rules of engagement.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Hostile Declaration = "hostile"
	// Furball indicates that non-friendly and friendly aircraft are inside of 5 nauctical miles of each other.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Furball Declaration = "furball"
	// Unable indications that the responder is unable to provide a declaration as requested.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Unable Declaration = "unable"
	// Clean indicates there is no sensor information on the contact.
	// Reference: ATP 1-02.1 Chapter I Table 2.
	Clean Declaration = "clean"
)

// DeclareRequest is a DECLARE call.
// Reference: ATP 3-52.4 Chapter V section 6.
type DeclareRequest struct {
	// Callsign of the friendly aircraft requesting DECLARE.
	Callsign string
	// Sour indicates if the player attempted a DECLARE request without
	// providing coordinates for the contact.
	Sour bool
	// IsBRAA indicates if the contact is provided using Bullseye (false) or BRAA (true).
	IsBRAA bool
	// IsAmbiguous indicates if the requestor did not explicitly state if they
	// were providing Bullseye or BRAA coordinates.
	IsAmbiguous bool
	// Bullseye of the contact, if provided using Bullseye.
	Bullseye Bullseye
	// Bearing of the contact, if provided using BRAA.
	Bearing bearings.Bearing
	/// Range to the contact, if provided using BRAA.
	Range unit.Length
	// Altitude of the contact above sea level, rounded to the nearest thousands of feet.
	Altitude unit.Length
	// Track direction. Optional, used to discriminate between multiple contacts at the same location.
	Track Track
}

func (r DeclareRequest) String() string {
	s := fmt.Sprintf("DECLARE for %s: ", r.Callsign)
	if r.Sour {
		s += "No coordinates provided"
		return s
	}
	if r.IsBRAA {
		s += fmt.Sprintf("bearing %s, range %.0f", r.Bearing, r.Range.NauticalMiles())
		if r.Altitude != 0 {
			s += fmt.Sprintf(", altitude %.0f", r.Altitude.Feet())
		}
	} else {
		s += fmt.Sprintf("bullseye %s", &r.Bullseye)
	}
	if r.Track != UnknownDirection {
		s += fmt.Sprintf(", track %s", r.Track)
	}
	return s
}

// DeclareResponse is a response to a DECLARE call.
// Reference: ATP 3-52.4 Chapter V section 6.
type DeclareResponse struct {
	// Callsign of the friendly aircraft requesting DECLARE.
	Callsign string
	// Sour indicates if the controller is unable to provide a declaration
	// because the friendly aircraft did not provide coordinates for the
	// contact.
	Sour bool
	// If readback is not nil, the controller should read back the coordinate
	// in the response.
	Readback *Bullseye
	// Declaration of the contact.
	Declaration Declaration
	// Group that was identified, if a specific one was identifiable.
	// This may be nil if Declaration is Furball, Unable, or Clean.
	Group Group
}
