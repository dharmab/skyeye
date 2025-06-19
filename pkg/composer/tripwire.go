package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeTripwireResponse constructs natural language brevity for educating a caller about threat monitoring.
func (c *Composer) ComposeTripwireResponse(response brevity.TripwireResponse) NaturalLanguageResponse {
	reply := c.composeCallsigns(response.Callsign) + ", I'm not OverlordBot. Tripwires are not real brevity per the MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication. If your name is set correctly in-game and in SRS, I'll automatically monitor you on the radar and provide threat warnings. You can also send requests such as picture, bogey dope, snaplock, spiked, declare and alpha check."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
