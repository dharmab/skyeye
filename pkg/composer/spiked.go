package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSpikedResponse implements [Composer.ComposeSpikedResponse].
func (c *composer) ComposeSpikedResponse(response brevity.SpikedResponse) NaturalLanguageResponse {
	if response.Status {
		reply := fmt.Sprintf("%s, spike range %d, %d, %s", response.Callsign, int(response.Range.NauticalMiles()), int(response.Altitude.Feet()), response.Aspect)
		isCardinalAspect := slices.Contains([]brevity.Aspect{brevity.Flank, brevity.Beam, brevity.Drag}, response.Aspect)
		isTrackKnown := response.Track != brevity.UnknownDirection
		if isCardinalAspect && isTrackKnown {
			reply = fmt.Sprintf("%s %s", reply, response.Track)
		}
		reply = fmt.Sprintf("%s, %s", reply, response.Declaration)
		if response.Contacts == 1 {
			reply = fmt.Sprintf("%s, single contact.", reply)
		} else if response.Contacts > 1 {
			reply = fmt.Sprintf("%s, %d contacts.", reply, response.Contacts)
		}
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s clean %d.", response.Callsign, c.callsign, int(response.Bearing.Degrees())),
		Speech:   fmt.Sprintf("%s, %s clean %s", response.Callsign, c.callsign, PronounceBearing(int(response.Bearing.Degrees()))),
	}
}
