package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeSnaplockResponse implements [Composer.ComposeSnaplockResponse].
func (c *composer) ComposeSnaplockResponse(response brevity.SnaplockResponse) NaturalLanguageResponse {
	if response.Declaration == brevity.Hostile || response.Declaration == brevity.Friendly {
		info := c.ComposeCoreInformationFormat(response.Group)
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s", strings.ToUpper(response.Callsign), info.Subtitle),
			Speech:   fmt.Sprintf("%s, %s", strings.ToUpper(response.Callsign), info.Speech),
		}
	}

	reply := fmt.Sprintf("%s, %s", strings.ToUpper(response.Callsign), response.Declaration)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
