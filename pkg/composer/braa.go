package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeBRAA(braa brevity.BRAA) NaturalLanguageResponse {
	var aspect string
	if braa.Aspect() != brevity.UnknownAspect {
		aspect = string(braa.Aspect())
	}
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(
			"BRAA %d/%d, %d, %s",
			int(braa.Bearing().Degrees()),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			aspect,
		),
		Speech: fmt.Sprintf(
			"brah %s, %d, %d, %s",
			PronounceBearing(int(braa.Bearing().Degrees())),
			int(braa.Range().NauticalMiles()),
			int(braa.Altitude().Feet()),
			aspect,
		),
	}
}
