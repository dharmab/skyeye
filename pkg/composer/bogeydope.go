package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeBogeyDopeResponse implements [Composer.ComposeBogeyDopeResponse].
func (c *composer) ComposeBogeyDopeResponse(response brevity.BogeyDopeResponse) NaturalLanguageResponse {
	if response.Group == nil {
		reply := fmt.Sprintf("%s, %s", response.Callsign, brevity.Clean)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	info := c.ComposeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", response.Callsign, info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", response.Callsign, info.Speech),
	}
}
