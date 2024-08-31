package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeMergedCall(call brevity.MergedCall) NaturalLanguageResponse {
	callsignList := strings.Join(call.Callsigns, ", ")
	group := c.ComposeMergedWithGroup(call.Group)
	template := "%s, merged. %s"

	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf(template, callsignList, group.Subtitle),
		Speech:   fmt.Sprintf(template, callsignList, group.Speech),
	}
}
