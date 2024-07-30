package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposePictureResponse implements [Composer.ComposePictureResponse].
func (c *composer) ComposePictureResponse(r brevity.PictureResponse) []NaturalLanguageResponse {
	info := c.ComposeCoreInformationFormat(r.Count, r.Groups, true)

	responses := make([]NaturalLanguageResponse, len(info))
	responses[0] = NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.callsign, info[0].Subtitle),
		Speech:   fmt.Sprintf("%s, %s", c.callsign, info[0].Speech),
	}
	for i := 1; i < len(info); i++ {
		responses[i] = info[i]
	}
	return responses
}
