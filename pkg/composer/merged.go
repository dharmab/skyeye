package composer

import (
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeMergedCall(call brevity.MergedCall) NaturalLanguageResponse {
	s := strings.Join(call.Callsigns, ", ") + ", merged."
	return NaturalLanguageResponse{
		Subtitle: s,
		Speech:   s,
	}
}
