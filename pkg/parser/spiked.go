package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseSpiked(callsign string, scanner *bufio.Scanner) (*brevity.SpikedRequest, bool) {
	bearing, ok := parseBearing(scanner)
	if !ok {
		return nil, false
	}
	return &brevity.SpikedRequest{
		Callsign: callsign,
		Bearing:  bearing,
	}, true
}
