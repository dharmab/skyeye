package parser

import (
	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseStrobe(callsign string, stream *token.Stream) (*brevity.StrobeRequest, bool) {
	bearing, ok := parseBearing(stream)
	if !ok {
		return nil, false
	}
	return &brevity.StrobeRequest{
		Callsign: callsign,
		Bearing:  bearing,
	}, true
}
