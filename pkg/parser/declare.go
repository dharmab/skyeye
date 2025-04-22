package parser

import (
	"bufio"
	"unicode"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

func parseDeclare(callsign string, scanner *bufio.Scanner) (*brevity.DeclareRequest, bool) {
	var bullseye brevity.Bullseye
	var bearing bearings.Bearing
	var _range unit.Length
	var isBRAA bool
	isAmbiguous := true
	for {
		if scanner.Text() == "" {
			ok := scanner.Scan()
			if !ok {
				return &brevity.DeclareRequest{
					Callsign: callsign,
					Sour:     true,
				}, true
			}
			continue
		}

		// if text only contains digits, assume bullseye format
		isNumeric := true
		for _, r := range scanner.Text() {
			if !unicode.IsDigit(r) {
				isNumeric = false
				break
			}
		}
		if isNumeric {
			log.Debug().Str("text", scanner.Text()).Msg("found numeric token, assuming format bullseye")
			bullseye = parseBullseye(scanner)
			break
		}

		parsedAsBullseye := false
		for _, word := range bullseyeWords {
			if isSimilar(scanner.Text(), word) {
				log.Debug().Str("text", scanner.Text()).Msg("found bullseye token")
				isAmbiguous = false
				bullseye = parseBullseye(scanner)
				if bullseye == nil {
					return nil, false
				}
				parsedAsBullseye = true
				break
			}
		}
		if parsedAsBullseye {
			log.Debug().Float64("bearing", bullseye.Bearing().Degrees()).Float64("distance", bullseye.Distance().NauticalMiles()).Msg("parsed bullseye")
			break
		}

		for _, word := range braaWords {
			if isSimilar(scanner.Text(), word) {
				log.Debug().Str("text", scanner.Text()).Msg("found braa token")
				isAmbiguous = false
				scanner.Scan()
				b, extra, ok := parseBearing(scanner)
				if !ok {
					return nil, false
				}
				bearing = b
				scanner = prependToScanner(scanner, extra)
				r, ok := parseRange(scanner)
				if !ok {
					return nil, false
				}
				_range = r
				isBRAA = true
				break
			}
		}

		if isBRAA {
			log.Debug().Float64("bearing", bearing.Degrees()).Float64("range", _range.NauticalMiles()).Msg("parsed bearing and range")
			break
		}

		if ok := scanner.Scan(); !ok {
			return nil, false
		}
	}

	altitude, ok := parseAltitude(scanner)
	if ok {
		log.Debug().Int("altitude", int(altitude.Feet())).Msg("parsed altitude")
	}

	track := parseTrack(scanner)
	log.Debug().Stringer("track", track).Msg("parsed track")

	if isBRAA {
		return &brevity.DeclareRequest{
			Callsign:    callsign,
			Bearing:     bearing,
			Range:       _range,
			Altitude:    altitude,
			Track:       track,
			IsBRAA:      true,
			IsAmbiguous: false,
		}, true
	}
	if bullseye == nil {
		return nil, false
	}
	return &brevity.DeclareRequest{
		Callsign:    callsign,
		Bullseye:    bullseye,
		Altitude:    altitude,
		Track:       track,
		IsAmbiguous: isAmbiguous,
	}, true
}
