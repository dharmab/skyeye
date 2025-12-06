package parser

import (
	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
)

var bullseyeWords = []string{"bullseye", "bulls"}

func parseBullseye(stream *token.Stream) *brevity.Bullseye {
	// Skip bullseye keyword if present
	for _, word := range bullseyeWords {
		if isSimilar(stream.Text(), word) {
			stream.Advance()
			break
		}
	}

	b, ok := parseBearing(stream)
	if !ok {
		log.Debug().Msg("failed to parse bullseye bearing")
		return nil
	}

	r, ok := parseRange(stream)
	if !ok {
		log.Debug().Msg("failed to parse bullseye range")
		return nil
	}

	return brevity.NewBullseye(b, r)
}

var braaWords = []string{"bra", "brah", "braa"}

func parseBRA(stream *token.Stream) (brevity.BRA, bool) {
	// Skip BRAA keyword if present
	for _, word := range braaWords {
		if isSimilar(stream.Text(), word) {
			stream.Advance()
			break
		}
	}

	b, ok := parseBearing(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRA bearing")
		return nil, false
	}

	r, ok := parseRange(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRA range")
		return nil, false
	}

	a, ok := parseAltitude(stream)
	if !ok {
		log.Debug().Msg("failed to parse BRA altitude")
	}

	return brevity.NewBRA(b, r, a), true
}

// parseBearing parses a 3 digit magnetic bearing. Each digit should be
// individually pronounced. Zeroes must be prefixed to values below 100.
func parseBearing(stream *token.Stream) (bearings.Bearing, bool) {
	bearing := 0 * unit.Degree
	digitsParsed := 0

	for digitsParsed < 3 && !stream.AtEnd() {
		tokenText := stream.Text()
		if tokenText == "" {
			stream.Advance()
			continue
		}

		if !hasDigits(tokenText) {
			stream.Advance()
			continue
		}

		// Parse digits from current token
		charsConsumed := 0
		for _, char := range tokenText {
			if d, err := numwords.ParseInt(string(char)); err == nil {
				bearing = bearing*10 + unit.Degree*unit.Angle(d)
				digitsParsed++
				charsConsumed++
				if digitsParsed == 3 {
					// If there are more characters in this token, save them as partial
					if charsConsumed < len(tokenText) {
						stream.SetPartialToken(tokenText[charsConsumed:])
					} else {
						stream.Advance()
					}
					return bearings.NewMagneticBearing(bearing), true
				}
			}
		}

		stream.Advance()
	}

	return bearings.NewMagneticBearing(bearing), digitsParsed == 3
}

// parseRange parses a distance. The number must be pronounced as a whole cardinal number.
func parseRange(stream *token.Stream) (unit.Length, bool) {
	if stream.AtEnd() {
		return 0, false
	}

	// Skip optional "for" keyword
	if isSimilar(stream.Text(), "for") {
		stream.Advance()
	}

	if stream.AtEnd() {
		return 0, false
	}

	d, ok := parseNaturalNumber(stream)
	if !ok {
		return 0, false
	}

	return unit.Length(d) * unit.NauticalMile, true
}

func parseAltitude(stream *token.Stream) (unit.Length, bool) {
	if stream.AtEnd() {
		return 0, false
	}

	// Skip optional altitude keywords
	text := stream.Text()
	if isSimilar(text, "at") || isSimilar(text, "altitude") || isSimilar(text, "angels") {
		stream.Advance()
	}

	if stream.AtEnd() {
		return 0, false
	}

	d, ok := parseNaturalNumber(stream)
	if !ok {
		return 0, false
	}

	altitude := unit.Length(d) * unit.Foot
	// Values below 100 are likely a player incorrectly saying "angels XX" instead of thousands of feet.
	if d < 100 {
		altitude = altitude * 1000
	}
	return altitude, true
}

func parseTrack(stream *token.Stream) brevity.Track {
	if stream.AtEnd() {
		return brevity.UnknownDirection
	}

	// Skip "track" keyword if present
	for isSimilar(stream.Text(), "track") {
		if !stream.Advance() {
			return brevity.UnknownDirection
		}
	}

	if stream.AtEnd() {
		return brevity.UnknownDirection
	}

	switch stream.Text() {
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

func parseNaturalNumber(stream *token.Stream) (int, bool) {
	s := stream.Text()
	d, err := numwords.ParseInt(s)
	if err != nil {
		log.Error().Err(err).Str("text", s).Msg("failed to parse natural number")
		return 0, false
	}
	stream.Advance()
	return d, true
}

// parseBearingOnly is a helper for requests that only require a bearing
// (e.g., SPIKED, STROBE). Returns the parsed bearing and success status.
func parseBearingOnly(stream *token.Stream) (bearings.Bearing, bool) {
	bearing, ok := parseBearing(stream)
	if !ok {
		log.Debug().Msg("failed to parse bearing")
		return bearings.NewMagneticBearing(0), false
	}
	return bearing, true
}
