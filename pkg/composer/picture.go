package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposePictureResponse(r brevity.PictureResponse) NaturalLanguageResponse {
	g := c.ComposeCoreInformationFormat(r.Groups())
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.callsign, g.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", c.callsign, g.Speech),
	}
}
