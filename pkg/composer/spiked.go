package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeSpikedResponse(r brevity.SpikedResponse) NaturalLanguageResponse {
	if r.Status {
		reply := fmt.Sprintf("%s, spike range %d, %d, %s", r.Callsign, int(r.Range.NauticalMiles()), int(r.Altitude.Feet()), r.Aspect)
		isCardinalAspect := slices.Contains([]brevity.Aspect{brevity.Flank, brevity.Beam, brevity.Drag}, r.Aspect)
		isTrackKnown := r.Track != brevity.UnknownDirection
		if isCardinalAspect && isTrackKnown {
			reply = fmt.Sprintf("%s %s", reply, r.Track)
		}
		reply = fmt.Sprintf("%s, %s", reply, r.Declaration)
		if r.Contacts == 1 {
			reply = fmt.Sprintf("%s, single contact.", reply)
		} else if r.Contacts > 1 {
			reply = fmt.Sprintf("%s, %d contacts.", reply, r.Contacts)
		}
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s clean %d.", r.Callsign, c.callsign, int(r.Bearing.Degrees())),
		Speech:   fmt.Sprintf("%s, %s clean %s", r.Callsign, c.callsign, PronounceBearing(int(r.Bearing.Degrees()))),
	}
}
