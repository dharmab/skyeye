package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBRAA constructs natural language brevity for communicating BRAA information.
func (c *composer) ComposeBRAA(braa brevity.BRAA, declaration brevity.Declaration) NaturalLanguageResponse {
	if !braa.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", braa.Bearing()).Msg("bearing provided to ComposeBRAA should be magnetic")
	}
	bearing := PronounceBearing(braa.Bearing())
	var aspect string
	if braa.Aspect() != brevity.UnknownAspect {
		aspect = string(braa.Aspect())
	}
	_range := int(braa.Range().NauticalMiles())
	altitude := c.ComposeAltitude(braa.Altitude(), declaration)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("BRAA %s/%d, %s, %s", braa.Bearing().String(), _range, altitude, aspect),
		Speech:   fmt.Sprintf("BRAA %s, %d, %s, %s", bearing, _range, altitude, aspect),
	}
}
