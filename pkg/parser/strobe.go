package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseStrobe(callsign string, scanner *bufio.Scanner) (*brevity.StrobeRequest, bool) {
	bearing, _, ok := parseBearing(scanner)
	if !ok {
		return nil, false
	}
	return &brevity.StrobeRequest{
		Callsign: callsign,
		Bearing:  bearing,
	}, true
}
