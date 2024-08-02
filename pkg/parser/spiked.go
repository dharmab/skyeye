package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (p *parser) parseSpiked(callsign string, scanner *bufio.Scanner) (*brevity.SpikedRequest, bool) {
	bearing, ok := p.parseBearing(scanner)
	if !ok {
		return nil, false
	}
	return &brevity.SpikedRequest{
		Callsign: callsign,
		Bearing:  bearing,
	}, true
}
