package composer

import (
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeMergedCall(call brevity.MergedCall) NaturalLanguageResponse {
	callsignList := strings.ToUpper(strings.Join(call.Callsigns, ", "))
	reply := callsignList + ", merged."
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
