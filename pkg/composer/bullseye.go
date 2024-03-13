package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeBullseye(bullseye brevity.Bullseye) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %d/%d",
			int(bullseye.Bearing().Degrees()),
			int(bullseye.Distance().NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			PronounceInt(int(bullseye.Bearing().Degrees())),
			int(bullseye.Distance().NauticalMiles()),
		),
	}
}
