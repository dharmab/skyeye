package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSpikedResponse constructs natural language brevity for responding to a SPIKED call.
func (c *Composer) ComposeSpikedResponse(response brevity.SpikedResponse) NaturalLanguageResponse {
	if response.Status {
		reply := fmt.Sprintf(
			"%s, spike range %d, %s, %s",
			c.composeCallsigns(response.Callsign),
			int(response.Range.NauticalMiles()),
			c.composeAltitude(response.Altitude, brevity.Bogey),
			response.Aspect)
		isCardinalAspect := slices.Contains([]brevity.Aspect{brevity.Flank, brevity.Beam, brevity.Drag}, response.Aspect)
		isTrackKnown := response.Track != brevity.UnknownDirection
		if isCardinalAspect && isTrackKnown {
			reply = fmt.Sprintf("%s %s", reply, response.Track)
		}
		reply = fmt.Sprintf("%s, %s", reply, response.Declaration)
		if response.Contacts == 1 {
			reply += ", single contact."
		} else if response.Contacts > 1 {
			reply = fmt.Sprintf("%s, %d contacts.", reply, response.Contacts)
		}
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	if response.Bearing == nil {
		message := fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), brevity.Unable)
		return NaturalLanguageResponse{
			Subtitle: message,
			Speech:   message,
		}
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s clean %d.", c.composeCallsigns(response.Callsign), c.composeCallsigns(c.Callsign), int(response.Bearing.Degrees())),
		Speech:   fmt.Sprintf("%s, %s, clean - %s", c.composeCallsigns(response.Callsign), c.composeCallsigns(c.Callsign), pronounceBearing(response.Bearing)),
	}
}
