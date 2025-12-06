package parser

import (
	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseSnaplock(callsign string, stream *token.Stream) (*brevity.SnaplockRequest, bool) {
	bra, ok := parseBRA(stream)
	if !ok {
		return nil, false
	}
	return &brevity.SnaplockRequest{
		Callsign: callsign,
		BRA:      bra,
	}, true
}
