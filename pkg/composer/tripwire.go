package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeTripwireResponse constructs natural language brevity for educating a caller about threat monitoring.
func (c *Composer) ComposeTripwireResponse(response brevity.TripwireResponse) NaturalLanguageResponse {
	reply := c.composeCallsigns(response.Callsign) + ", I am not OverlordBot. I am a GCI bot called sky eye which implements the real-world MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication. If your SRS client name and in-game name are set correctly, I automatically provide threat warnings based on hostile weapon system capabilities and the briefed minimum threat radius configured by the server administrator. In other words, you should not set a threat radius. Instead, I am monitoring the radar and I will warn you about any threats which require your attention. For more information, please read the player guide which you can find online by searching for sky eye GCI player guide."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
