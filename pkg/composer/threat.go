package composer

import (
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

// ComposeThreatCall implements [Composer.ComposeThreatCall].
func (c *composer) ComposeThreatCall(call brevity.ThreatCall) NaturalLanguageResponse {
	group := c.ComposeGroup(call.Group)
	callsignList := c.ComposeCallsigns(call.Callsigns...)
	return NaturalLanguageResponse{
		Subtitle: fmt.Sprintf("%s, %s", callsignList, strings.ToLower(group.Subtitle[:1])+group.Subtitle[1:]),
		Speech:   fmt.Sprintf("%s, %s", callsignList, group.Speech),
	}
}
