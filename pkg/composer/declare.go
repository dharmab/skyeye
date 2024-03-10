package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposeDeclareResponse(brevity.DeclareResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "DECLARE not yet implemented",
		Speech:   "Sorry, I don't know how to respond to DECLARE yet. I'm still learning!",
	}
}
