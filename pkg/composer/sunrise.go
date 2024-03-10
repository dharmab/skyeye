package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeSunriseCall(brevity.SunriseCall) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "SUNRISE not yet implemented",
		Speech:   "Sorry, I don't know how to make a SUNRISE call. I'm still learning!",
	}
}
