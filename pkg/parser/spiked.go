package parser

import (
	"bufio"
	"math"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

type spikedRequest struct {
	callsign       string
	bearingDegrees int
}

var _ brevity.SpikedRequest = &spikedRequest{}

func (r *spikedRequest) Spiked() {}

func (r *spikedRequest) Callsign() string {
	return r.callsign
}

func (r *spikedRequest) Bearing() unit.Angle {
	return unit.Angle(r.bearingDegrees) * unit.Degree
}

func (p *parser) parseSpiked(callsign string, scanner *bufio.Scanner) (brevity.SpikedRequest, bool) {
	bearing, ok := p.parseBearing(scanner)
	if !ok {
		return nil, false
	}
	return &spikedRequest{
		callsign:       callsign,
		bearingDegrees: int(math.Round(bearing.Degrees())),
	}, true
}
