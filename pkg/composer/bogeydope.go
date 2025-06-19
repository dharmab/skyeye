package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBogeyDopeResponse constructs natural language brevity for responding to a BOGEY DOPE call.
func (c *Composer) ComposeBogeyDopeResponse(response brevity.BogeyDopeResponse) NaturalLanguageResponse {
	if response.Group == nil {
		reply := fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), brevity.Clean)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	if !response.Group.BRAA().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", response.Group.BRAA().Bearing()).Msg("bearing provided to ComposeBogeyDopeResponse should be magnetic")
	}
	info := c.composeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), lowerFirst(info.Speech)),
		Speech:   fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), info.Speech),
	}
}
