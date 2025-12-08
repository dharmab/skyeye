package parser

import (
	"strings"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

var bogeyFilterMap = map[string]brevity.ContactCategory{
	"airplane":    brevity.FixedWing,
	"planes":      brevity.FixedWing,
	"fighter":     brevity.FixedWing,
	"fixed wing":  brevity.FixedWing,
	"helicopter":  brevity.RotaryWing,
	"chopper":     brevity.RotaryWing,
	"helo":        brevity.RotaryWing,
	"rotary wing": brevity.RotaryWing,
}

func parseBogeyDope(callsign string, stream *token.Stream) (*brevity.BogeyDopeRequest, bool) {
	filter := brevity.Aircraft

	remainingText := stream.RemainingText()

	for k, v := range bogeyFilterMap {
		if strings.Contains(remainingText, k) {
			filter = v
			break
		}
	}

	return &brevity.BogeyDopeRequest{Callsign: callsign, Filter: filter}, true
}
