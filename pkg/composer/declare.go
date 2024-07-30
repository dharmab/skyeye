package composer

import (
	"fmt"
	"slices"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeDeclareResponse implements [Composer.ComposeDeclareResponse].
func (c *composer) ComposeDeclareResponse(r brevity.DeclareResponse) NaturalLanguageResponse {
	if slices.Contains([]brevity.Declaration{brevity.Furball, brevity.Unable, brevity.Clean}, r.Declaration) {
		return NaturalLanguageResponse{
			Subtitle: fmt.Sprintf("%s, %s.", r.Callsign, r.Declaration),
			Speech:   fmt.Sprintf("%s, %s", r.Callsign, r.Declaration),
		}
	}
	g := c.ComposeCoreInformationFormat(1, []brevity.Group{r.Group}, false)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", r.Callsign, g[0].Subtitle),
		Speech:   fmt.Sprintf("%s, %s", r.Callsign, g[0].Speech),
	}
}
