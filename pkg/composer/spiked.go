package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeSpikedResponse(r brevity.SpikedResponse) NaturalLanguageResponse {
	if r.Status {
		reply := fmt.Sprintf("%s, spike range %d, %d, %s, %s", r.Callsign, int(r.Range.NauticalMiles()), int(r.Altitude.Feet()), r.Aspect, r.Declaration)
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
