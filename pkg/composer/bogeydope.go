package composer

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeBogeyDopeResponse(r brevity.BogeyDopeResponse) NaturalLanguageResponse {
	if r.Group == nil {
		reply := fmt.Sprintf("%s, %s", r.Callsign, brevity.Clean)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	groups := c.ComposeCoreInformationFormat([]brevity.Group{r.Group})
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", r.Callsign, groups.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", r.Callsign, groups.Speech),
	}
}
