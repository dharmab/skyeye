package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// ComposeDeclareResponse implements [Composer.ComposeDeclareResponse].
func (c *composer) ComposeDeclareResponse(response brevity.DeclareResponse) NaturalLanguageResponse {
	if !response.Group.BRAA().Bearing().IsMagnetic() {
		log.Error().Any("bearing", response.Group.BRAA().Bearing()).Msg("bearing provided to ComposeDeclareResponse should be magnetic")
	}
	if slices.Contains([]brevity.Declaration{brevity.Furball, brevity.Unable, brevity.Clean}, response.Declaration) {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", response.Callsign, response.Declaration),
			Speech:   fmt.Sprintf("%s, %s", response.Callsign, response.Declaration),
		}
	}
	info := c.ComposeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", response.Callsign, info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", response.Callsign, info.Speech),
	}
}
