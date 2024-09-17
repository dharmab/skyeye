package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
)

var bullseyeWords = []string{"bullseye", "bulls"}

func (p *parser) parseBullseye(scanner *bufio.Scanner) *brevity.Bullseye {
	skipWords(scanner, bullseyeWords...)

	b, ok := p.parseBearing(scanner)
	if !ok {
		return nil
	}

	if !skipWords(scanner, "for") {
		return nil
	}

	r, ok := p.parseRange(scanner)
	if !ok {
		return nil
	}

	return brevity.NewBullseye(b, r)
}

var braaWords = []string{"bra", "brah", "braa"}

func (p *parser) parseBRA(scanner *bufio.Scanner) (brevity.BRA, bool) {
	if !skipWords(scanner, braaWords...) {
		return nil, false
	}
	b, ok := p.parseBearing(scanner)
	if !ok {
		return nil, false
	}

	for scanner.Text() == "for" {
		ok := scanner.Scan()
		if !ok {
			return nil, false
		}
	}

	r, ok := p.parseRange(scanner)
	if !ok {
		return nil, false
	}

	a, ok := p.parseAltitude(scanner)
	if !ok {
		return nil, false
	}

	return brevity.NewBRA(b, r, a), true
}

// parseBearing parses a 3 digit magnetic bearing. Each digit must be individually pronounced. Zeroes must be prefixed to values below 100.
func (p *parser) parseBearing(scanner *bufio.Scanner) (bearings.Bearing, bool) {
	bearing := 0 * unit.Degree
	digitsParsed := 0
	for digitsParsed < 3 {
		for _, char := range scanner.Text() {
			if d, err := numwords.ParseInt(string(char)); err == nil {
				bearing = bearing*10 + unit.Degree*unit.Angle(d)
				digitsParsed++
			}
			if digitsParsed == 3 {
				return bearings.NewMagneticBearing(bearing), true
			}
		}
		ok := scanner.Scan()
		if !ok {
			return bearings.NewMagneticBearing(bearing), true
		}
	}
	return bearings.NewMagneticBearing(0), false
}

// parseRange parses a distance. The number must be pronounced as a whole cardinal number.
func (p *parser) parseRange(scanner *bufio.Scanner) (unit.Length, bool) {
	if !scanner.Scan() {
		return 0, false
	}
	if !skipWords(scanner, "for") {
		return 0, false
	}
	d, ok := p.parseNaturalNumber(scanner)
	if !ok {
		return 0, false
	}
	return unit.Length(d) * unit.NauticalMile, true
}

func (p *parser) parseAltitude(scanner *bufio.Scanner) (unit.Length, bool) {
	if !scanner.Scan() {
		return 0, false
	}
	if !skipWords(scanner, "at", "altitude") {
		return 0, false
	}
	d, ok := p.parseNaturalNumber(scanner)
	if !ok {
		return 0, false
	}
	return unit.Length(d) * unit.Foot, true
}

func (p *parser) parseTrack(scanner *bufio.Scanner) brevity.Track {
	for scanner.Text() == "track" {
		ok := scanner.Scan()
		if !ok {
			return brevity.UnknownDirection
		}
	}

	switch scanner.Text() {
	case "north":
		return brevity.North
	case "northeast":
		return brevity.Northeast
	case "east":
		return brevity.East
	case "southeast":
		return brevity.Southeast
	case "south":
		return brevity.South
	case "southwest":
		return brevity.Southwest
	case "west":
		return brevity.West
	case "northwest":
		return brevity.Northwest
	default:
		return brevity.UnknownDirection
	}
}

func (p *parser) parseNaturalNumber(scanner *bufio.Scanner) (int, bool) {
	s := scanner.Text()
	d, err := numwords.ParseInt(s)
	if err != nil {
		log.Error().Err(err).Str("text", s).Msg("failed to parse natural number")
		return 0, false
	}
	return d, true
}
