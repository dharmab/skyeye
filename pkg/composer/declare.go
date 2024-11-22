package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeDeclareResponse implements [Composer.ComposeDeclareResponse].
func (c *composer) ComposeDeclareResponse(response brevity.DeclareResponse) NaturalLanguageResponse {
	if slices.Contains([]brevity.Declaration{brevity.Furball, brevity.Unable, brevity.Clean}, response.Declaration) {
		reply := fmt.Sprintf("%s, %s.", c.ComposeCallsigns(response.Callsign), response.Declaration)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	info := c.ComposeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.ComposeCallsigns(response.Callsign), info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", c.ComposeCallsigns(response.Callsign), info.Speech),
	}
}
