package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeBogeyDopeResponse implements [Composer.ComposeBogeyDopeResponse].
func (c *composer) ComposeBogeyDopeResponse(r brevity.BogeyDopeResponse) NaturalLanguageResponse {
	if r.Group == nil {
		reply := fmt.Sprintf("%s, %s", r.Callsign, brevity.Clean)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	response := c.ComposeCoreInformationFormat(r.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", r.Callsign, response.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", r.Callsign, response.Speech),
	}
}
