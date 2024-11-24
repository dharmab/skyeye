package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeBRAA constructs natural language brevity for communicating BRAA information.
func (c *composer) ComposeBRAA(braa brevity.BRAA, declaration brevity.Declaration) NaturalLanguageResponse {
	if !braa.Bearing().IsMagnetic() {
		log.Error().Stringer("bearing", braa.Bearing()).Msg("bearing provided to ComposeBRAA should be magnetic")
	}
	bearing := PronounceBearing(braa.Bearing())
	_range := int(braa.Range().NauticalMiles())
	altitude := c.ComposeAltitude(braa.Altitude(), declaration)
	resp := NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("BRAA %s/%d, %s", braa.Bearing().String(), _range, altitude),
		Speech:   fmt.Sprintf("BRAA %s, %d, %s", bearing, _range, altitude),
	}

	isAspectKnown :=
		braa.Aspect() != brevity.UnknownAspect && !slices.Contains([]brevity.Declaration{
			brevity.Furball,
			brevity.Unable,
			brevity.Clean,
		}, declaration)
	if isAspectKnown {
		aspect := fmt.Sprintf(", aspect %s", braa.Aspect())
		resp.Speech += aspect
		resp.Subtitle += aspect
	}

	return resp
}
