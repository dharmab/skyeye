package parser

import (
	"unicode"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func parseDeclare(callsign string, stream *token.Stream) (*brevity.DeclareRequest, bool) {
	if stream.AtEnd() {
		return &brevity.DeclareRequest{
			Callsign: callsign,
			Sour:     true,
		}, true
	}

	for !stream.AtEnd() {
		text := stream.Text()

		for _, word := range braaWords {
			if isSimilar(text, word) {
				log.Debug().Str("text", text).Msg("found BRAA keyword")
				stream.Advance()
				return parseDeclareAsBRAA(callsign, stream)
			}
		}

		for _, word := range bullseyeWords {
			if isSimilar(text, word) {
				log.Debug().Str("text", text).Msg("found bullseye keyword")
				stream.Advance()
				return parseDeclareAsBullseye(callsign, stream, false)
			}
		}

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

		stream.Advance()
	}

	log.Debug().Msg("unable to determine declare format")
	return nil, false
}

func parseDeclareAsBRAA(callsign string, stream *token.Stream) (*brevity.DeclareRequest, bool) {
	bearing, ok := parseBearing(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRAA bearing")
		return nil, false
	}

	rng, ok := parseRange(stream)
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
		Range:       rng,
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
