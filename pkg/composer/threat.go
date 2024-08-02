package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeThreatCall implements [Composer.ComposeThreatCall].
func (c *composer) ComposeThreatCall(call brevity.ThreatCall) NaturalLanguageResponse {
	if !call.Group.BRAA().Bearing().IsMagnetic() {
		log.Error().Any("bearing", call.Group.BRAA().Bearing()).Msg("bearing provided to ComposeThreatCall should be magnetic")
	}
	group := c.ComposeGroup(call.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", call.Callsign, group.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", call.Callsign, group.Speech),
	}
}
