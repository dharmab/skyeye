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

func parseBullseye(scanner *bufio.Scanner) brevity.Bullseye {
	if !skipWords(scanner, bullseyeWords...) {
		return nil
	}

	b, extra, ok := parseBearing(scanner)
	if !ok {
		log.Debug().Msg("failed to parse bullseye bearing")
		return nil
	}
	scanner = prependToScanner(scanner, extra)

	r, ok := parseRange(scanner)
	if !ok {
		log.Debug().Msg("failed to parse bullseye range")
		return nil
	}

	return brevity.NewBullseye(b, r)
}

var braaWords = []string{"bra", "brah", "braa"}

func parseBRA(scanner *bufio.Scanner) (brevity.BRA, bool) {
	if !skipWords(scanner, braaWords...) {
		return nil, false
	}
	b, extra, ok := parseBearing(scanner)
	if !ok {
		log.Debug().Msg("failed to parse BRA bearing")
		return nil, false
	}
	scanner = prependToScanner(scanner, extra)

	r, ok := parseRange(scanner)
	if !ok {
		log.Debug().Msg("failed to parse BRA range")
		return nil, false
	}

	a, ok := parseAltitude(scanner)
	if !ok {
		log.Debug().Msg("failed to parse BRA altitude")
	}

	return brevity.NewBRA(b, r, a), true
}

// parseBearing parses a 3 digit magnetic bearing. Each digit should be
// individually pronounced. Zeroes must be prefixed to values below 100.
// Returns the bearing (if successfully parsed), the remaining text if any
// characters were left in a token after parsing 3 digits, and a boolean
// indicating if the bearing was successfully parsed.
func parseBearing(scanner *bufio.Scanner) (bearings.Bearing, string, bool) {
	bearing := 0 * unit.Degree
	digitsParsed := 0
	for digitsParsed < 3 {
		token := scanner.Text()
		if !hasDigits(token) {
			if !scanner.Scan() {
				return bearings.NewMagneticBearing(0), "", false
			}
			continue
		}
		extra := token
		for _, char := range token {
			if d, err := numwords.ParseInt(string(char)); err == nil {
				bearing = bearing*10 + unit.Degree*unit.Angle(d)
				digitsParsed++
			}
			extra = extra[1:]
			if digitsParsed == 3 {
				return bearings.NewMagneticBearing(bearing), extra, true
			}
		}
		if !scanner.Scan() {
			return bearings.NewMagneticBearing(bearing), extra, true
		}
	}
	return bearings.NewMagneticBearing(0), "", false
}

// parseRange parses a distance. The number must be pronounced as a whole cardinal number.
func parseRange(scanner *bufio.Scanner) (unit.Length, bool) {
	if !scanner.Scan() {
		return 0, false
	}
	if !skipWords(scanner, "for") {
		return 0, false
	}
	d, ok := parseNaturalNumber(scanner)
	if !ok {
		return 0, false
	}
	return unit.Length(d) * unit.NauticalMile, true
}

func parseAltitude(scanner *bufio.Scanner) (unit.Length, bool) {
	if !scanner.Scan() {
		return 0, false
	}
	if !skipWords(scanner, "at", "altitude", "angels") {
		return 0, false
	}
	d, ok := parseNaturalNumber(scanner)
	if !ok {
		return 0, false
	}

	altitude := unit.Length(d) * unit.Foot
	// Values below 100 are likely a player incorrectly saying "angels XX" intead of thousands of feet.
	if d < 100 {
		altitude = altitude * 1000
	}
	return altitude, true
}

func parseTrack(scanner *bufio.Scanner) brevity.Track {
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

func parseNaturalNumber(scanner *bufio.Scanner) (int, bool) {
	s := scanner.Text()
	d, err := numwords.ParseInt(s)
	if err != nil {
		log.Error().Err(err).Str("text", s).Msg("failed to parse natural number")
		return 0, false
	}
	return d, true
}
