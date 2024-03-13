package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeBullseye(r brevity.Bullseye) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"bullseye %d/%d",
			int(r.Bearing().Degrees()),
			int(r.Distance().NauticalMiles()),
		),
		Speech: fmt.Sprintf(
			"bullseye %s, %d",
			PronounceInt(int(r.Bearing().Degrees())),
			int(r.Distance().NauticalMiles()),
		),
	}
}
