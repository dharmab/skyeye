package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

type declareRequest struct {
	callsign string
	bullseye brevity.Bullseye
	altitude unit.Length
	hasTrack bool
	track    brevity.Track
}

var _ brevity.DeclareRequest = &declareRequest{}

func (r *declareRequest) Callsign() string {
	return r.callsign
}

func (r *declareRequest) Location() brevity.Bullseye {
	return r.bullseye
}

func (r *declareRequest) Altitude() unit.Length {
	return r.altitude
}

func (r *declareRequest) Track() brevity.Track {
	return r.track
}

func (r *declareRequest) HasTrack() bool {
	return r.hasTrack
}

func (p *parser) parseDeclare(callsign string, scanner *bufio.Scanner) (brevity.DeclareRequest, bool) {
	bullseye, ok := p.parseBullseye(scanner)
	if !ok {
		return nil, false
	}
	altitude, ok := p.parseAltitude(scanner)
	if !ok {
		return nil, false
	}
	track, trackOk := p.parseTrack(scanner)

	return &declareRequest{
		callsign: callsign,
		bullseye: bullseye,
		altitude: altitude,
		hasTrack: trackOk,
		track:    track,
	}, true
}
