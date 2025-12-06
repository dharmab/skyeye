package parser

import (
	"unicode"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func parseDeclare(callsign string, stream *token.Stream) (*brevity.DeclareRequest, bool) {
	// If no data at all, return sour declare
	if stream.AtEnd() {
		return &brevity.DeclareRequest{
			Callsign: callsign,
			Sour:     true,
		}, true
	}

	// Search for format keywords, skipping filler words
	for !stream.AtEnd() {
		text := stream.Text()

		// Explicit BRAA format
		for _, word := range braaWords {
			if isSimilar(text, word) {
				log.Debug().Str("text", text).Msg("found BRAA keyword")
				stream.Advance()
				return parseDeclareAsBRAA(callsign, stream)
			}
		}

		// Explicit Bullseye format
		for _, word := range bullseyeWords {
			if isSimilar(text, word) {
				log.Debug().Str("text", text).Msg("found bullseye keyword")
				stream.Advance()
				return parseDeclareAsBullseye(callsign, stream, false)
			}
		}

		// Implicit bullseye format (starts with digits, ambiguous)
		isNumeric := true
		for _, r := range text {
			if !unicode.IsDigit(r) {
				isNumeric = false
				break
			}
		}

		if isNumeric {
			log.Debug().Str("text", text).Msg("found numeric token, assuming bullseye format")
			return parseDeclareAsBullseye(callsign, stream, true)
		}

		// Skip this filler word and try next token
		stream.Advance()
	}

	// No recognized format found
	log.Debug().Msg("unable to determine declare format")
	return nil, false
}

func parseDeclareAsBRAA(callsign string, stream *token.Stream) (*brevity.DeclareRequest, bool) {
	bearing, ok := parseBearing(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRAA bearing")
		return nil, false
	}

	rang, ok := parseRange(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRAA range")
		return nil, false
	}

	altitude, ok := parseAltitude(stream)
	if ok {
		log.Debug().Int("altitude", int(altitude.Feet())).Msg("parsed altitude")
	}

	track := parseTrack(stream)
	log.Debug().Stringer("track", track).Msg("parsed track")

	return &brevity.DeclareRequest{
		Callsign:    callsign,
		Bearing:     bearing,
		Range:       rang,
		Altitude:    altitude,
		Track:       track,
		IsBRAA:      true,
		IsAmbiguous: false,
	}, true
}

func parseDeclareAsBullseye(callsign string, stream *token.Stream, isAmbiguous bool) (*brevity.DeclareRequest, bool) {
	bullseye := parseBullseye(stream)
	if bullseye == nil {
		log.Debug().Msg("failed to parse bullseye")
		return nil, false
	}

	log.Debug().
		Float64("bearing", bullseye.Bearing().Degrees()).
		Float64("distance", bullseye.Distance().NauticalMiles()).
		Msg("parsed bullseye")

	altitude, ok := parseAltitude(stream)
	if ok {
		log.Debug().Int("altitude", int(altitude.Feet())).Msg("parsed altitude")
	}

	track := parseTrack(stream)
	log.Debug().Stringer("track", track).Msg("parsed track")

	return &brevity.DeclareRequest{
		Callsign:    callsign,
		Bullseye:    *bullseye,
		Altitude:    altitude,
		Track:       track,
		IsAmbiguous: isAmbiguous,
	}, true
}
