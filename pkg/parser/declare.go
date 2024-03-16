package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (p *parser) parseDeclare(callsign string, scanner *bufio.Scanner) (*brevity.DeclareRequest, bool) {
	bullseye := p.parseBullseye(scanner)
	if bullseye == nil {
		return nil, false
	}
	altitude, ok := p.parseAltitude(scanner)
	if !ok {
		return nil, false
	}
	track := p.parseTrack(scanner)

	return &brevity.DeclareRequest{
		Callsign: callsign,
		Location: *bullseye,
		Altitude: altitude,
		Track:    track,
	}, true
}
