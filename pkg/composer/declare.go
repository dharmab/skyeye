package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeDeclareResponse constructs natural language brevity for responding to a DECLARE call.
func (c *Composer) ComposeDeclareResponse(response brevity.DeclareResponse) NaturalLanguageResponse {
	if response.Sour {
		reply := fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), "unable, timber sour. Repeat your request with bullseye or BRAA position included.")
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}

	if slices.Contains([]brevity.Declaration{brevity.Furball, brevity.Unable, brevity.Clean}, response.Declaration) {
		reply := fmt.Sprintf("%s, %s.", c.composeCallsigns(response.Callsign), response.Declaration)
		return NaturalLanguageResponse{
			Subtitle: reply,
			Speech:   reply,
		}
	}
	info := c.composeCoreInformationFormat(response.Group)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), info.Subtitle),
		Speech:   fmt.Sprintf("%s, %s", c.composeCallsigns(response.Callsign), info.Speech),
	}
}
