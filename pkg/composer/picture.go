package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposePictureResponse constructs natural language brevity for responding to a PICTURE call.
func (c *Composer) ComposePictureResponse(response brevity.PictureResponse) NaturalLanguageResponse {
	info := c.composeCoreInformationFormat(response.Groups...)
	controllerCallsign := c.composeCallsigns(c.Callsign)
	if response.Count == 0 {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", controllerCallsign, brevity.Clean),
			Speech:   fmt.Sprintf("%s, %s", controllerCallsign, brevity.Clean),
		}
	}

	groupCountFillIn := "single group."
	if response.Count > 1 {
		groupCountFillIn = fmt.Sprintf("%d groups.", response.Count)
	}

	info.Speech = strings.TrimSpace(info.Speech)
	info.Subtitle = strings.TrimSpace(info.Subtitle)

	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s %s", controllerCallsign, groupCountFillIn, info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s %s", controllerCallsign, groupCountFillIn, info.Speech),
	}
}
