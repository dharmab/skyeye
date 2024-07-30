package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposePictureResponse implements [Composer.ComposePictureResponse].
func (c *composer) ComposePictureResponse(r brevity.PictureResponse) NaturalLanguageResponse {
	response := c.ComposeCoreInformationFormat(r.Groups...)
	if r.Count == 0 {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", c.callsign, brevity.Clean),
			Speech:   fmt.Sprintf("%s, %s", c.callsign, brevity.Clean),
		}
	}

	groupCountS := "single group."
	if r.Count > 1 {
		groupCountS = fmt.Sprintf("%d groups.", r.Count)
	}

	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s %s", c.callsign, groupCountS, response.Subtitle),
		Speech:   fmt.Sprintf("%s, %s %s", c.callsign, groupCountS, response.Speech),
	}
}
