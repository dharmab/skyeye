package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeBRAA(braa brevity.BRAA) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"BRAA %d/%d, %d, %s",
			int(braa.Bearing().Degrees()),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			braa.Aspect(),
		),
		Speech: fmt.Sprintf(
			"brah %s, %d, %d, %s",
			PronounceInt(int(braa.Bearing().Degrees())),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			braa.Aspect(),
		),
	}
}
