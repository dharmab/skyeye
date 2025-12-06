package parser

import (
	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseSpiked(callsign string, stream *token.Stream) (*brevity.SpikedRequest, bool) {
	bearing, ok := parseBearingOnly(stream)
	if !ok {
		return nil, false
	}
	return &brevity.SpikedRequest{
		Callsign: callsign,
		Bearing:  bearing,
	}, true
}
