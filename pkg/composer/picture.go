package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposePictureResponse implements [Composer.ComposePictureResponse].
func (c *composer) ComposePictureResponse(response brevity.PictureResponse) NaturalLanguageResponse {
	info := c.ComposeCoreInformationFormat(response.Groups...)
	if response.Count == 0 {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", c.callsign, brevity.Clean),
			Speech:   fmt.Sprintf("%s, %s", c.callsign, brevity.Clean),
		}
	}

	groupCountFillIn := "single group."
	if response.Count > 1 {
		groupCountFillIn = fmt.Sprintf("%d groups.", response.Count)
	}

	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s %s", c.callsign, groupCountFillIn, info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s %s", c.callsign, groupCountFillIn, info.Speech),
	}
}
