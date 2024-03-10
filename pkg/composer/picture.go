package composer

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *composer) ComposePictureResponse(brevity.PictureResponse) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: "PICTURE not yet implemented",
		Speech:   "Sorry, I don't know how to respond to PICTURE yet. I'm still learning!",
	}
}
