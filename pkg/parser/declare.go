package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (p *parser) parseDeclare(callsign string, scanner *bufio.Scanner) (*brevity.DeclareRequest, bool) {
	bullseye := p.parseBullseye(scanner)
	if bullseye == nil {
		return nil, false
	}
	log.Debug().Float64("bearing", bullseye.Bearing().Degrees()).Float64("distance", bullseye.Distance().NauticalMiles()).Msg("parsed bullseye")
	altitude, ok := p.parseAltitude(scanner)
	if !ok {
		return nil, false
	}
	log.Debug().Int("altitude", int(altitude.Feet())).Msg("parsed altitude")
	track := p.parseTrack(scanner)
	log.Debug().Str("track", string(track)).Msg("parsed track")

	return &brevity.DeclareRequest{
		Callsign: callsign,
		Location: *bullseye,
		Altitude: altitude,
		Track:    track,
	}, true
}
