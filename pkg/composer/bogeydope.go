package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBogeyDopeResponse implements [Composer.ComposeBogeyDopeResponse].
func (c *composer) ComposeBogeyDopeResponse(response brevity.BogeyDopeResponse) NaturalLanguageResponse {
	if response.Group == nil {
		reply := fmt.Sprintf("%s, %s", c.ComposeCallsigns(response.Callsign), brevity.Clean)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	if !response.Group.BRAA().Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", response.Group.BRAA().Bearing()).Msg("bearing provided to ComposeBogeyDopeResponse should be magnetic")
	}
	info := c.ComposeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.ComposeCallsigns(response.Callsign), info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", c.ComposeCallsigns(response.Callsign), info.Speech),
	}
}
