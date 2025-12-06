// Package parser converts brevity requests from natural language into structured forms.
package parser

import (
	"strings"

	"github.com/dharmab/skyeye/internal/parser/token"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
)

const (
	// maxInputLength is the maximum length of input text to process.
	// Prevents processing extremely long inputs that may indicate garbage data.
	maxInputLength = 1000
)

// Parser converts brevity requests from natural language into structured forms.
type Parser struct {
	controllerCallsign string
	enableTextLogging  bool
}

// New creates a new parser.
func New(callsign string, enableTextLogging bool) *Parser {
	return &Parser{
		controllerCallsign: strings.ReplaceAll(callsign, " ", ""),
		enableTextLogging:  enableTextLogging,
	}
}

// Anyface is a brevity codeword that can be used in place of a GCI callsign.
const Anyface string = "anyface"

const (
	alphaCheck string = "alpha"
	bogeyDope  string = "bogey"
	checkIn    string = "check in"
	declare    string = "declare"
	picture    string = "picture"
	radioCheck string = "radio"
	shopping   string = "shopping"
	snaplock   string = "snaplock"
	spiked     string = "spiked"
	strobe     string = "strobe"
	tripwire   string = "tripwire"
)

var requestWords = []string{radioCheck, alphaCheck, bogeyDope, declare, picture, spiked, strobe, snaplock, tripwire, shopping}

// findControllerCallsign searches for the GCI callsign in the given fields.
// Returns the heard callsign, remaining text after it, and whether it was found.
func (p *Parser) findControllerCallsign(fields []string) (heard string, rest string, ok bool) {
	for i := range fields {
		candidate := strings.Join(fields[:i+1], " ")
		for _, wakePhrase := range []string{p.controllerCallsign, Anyface} {
			if isSimilar(strings.TrimSpace(candidate), strings.ToLower(wakePhrase)) {
				ok = true
				heard = candidate
				rest = strings.Join(fields[i+1:], " ")
				return
			}
		}
	}
	return "", "", false
}

// handleNoRequestWord handles cases where we heard the GCI callsign and pilot callsign
// but couldn't identify a specific request word.
func handleNoRequestWord(tx, pilotCallsign string) any {
	if strings.Contains(tx, checkIn) {
		return &brevity.CheckInRequest{Callsign: pilotCallsign}
	}
	return &brevity.UnableToUnderstandRequest{Callsign: pilotCallsign}
}

// parseRequestWithArgs attempts to parse a request that requires additional arguments
// beyond the request word itself (e.g., BOGEY DOPE, DECLARE, SPIKED).
func parseRequestWithArgs(requestWord, pilotCallsign string, requestArgs []string) any {
	stream := token.New(strings.Join(requestArgs, " "))

	switch requestWord {
	case bogeyDope:
		if request, ok := parseBogeyDope(pilotCallsign, stream); ok {
			return request
		}
	case declare:
		if request, ok := parseDeclare(pilotCallsign, stream); ok {
			return request
		}
	case spiked:
		if request, ok := parseSpiked(pilotCallsign, stream); ok {
			return request
		}
	case strobe:
		if request, ok := parseStrobe(pilotCallsign, stream); ok {
			return request
		}
	case snaplock:
		if request, ok := parseSnaplock(pilotCallsign, stream); ok {
			return request
		}
	}

	return &brevity.UnableToUnderstandRequest{Callsign: pilotCallsign}
}

func findRequestWord(fields []string) (string, int, bool) {
	for i, field := range fields {
		field = strings.TrimPrefix(field, "request")
		for _, word := range requestWords {
			if isSimilar(word, field) {
				return word, i, true
			}
			// HACK: Also compare the first half of long fields separately.
			// Handles some cases of two words running together, e.g.
			// "bogeydope" instead of "bogey dope".
			if len(field) > halfFieldMinLength {
				halfField := field[:len(field)/2]
				if isSimilar(word, halfField) {
					return word, i, true
				}
			}
		}
	}
	return "", 0, false
}

// uncrushCallsign corrects a corner case where the GCI callsign and the
// following token have no space between them, e.g. "anyfaceeagle 1".
func (p *Parser) uncrushCallsign(s string) string {
	for _, callsign := range []string{p.controllerCallsign, Anyface} {
		lc := strings.ToLower(callsign)
		if strings.HasPrefix(s, lc) {
			return lc + " " + s[len(lc):]
		}
	}
	return s
}

// Parse reads natural language text, checks if it starts with the GCI
// callsign, and attempts to parse a request from the text. Returns a
// brevity request, or nil if the text does not start with the GCI
// callsign.
func (p *Parser) Parse(tx string) any {
	if tx == "" {
		return nil
	}
	if len(tx) > maxInputLength {
		log.Warn().Int("length", len(tx)).Msg("unusually long input truncated")
		tx = tx[:maxInputLength]
	}

	logger := log.With().Str("gci", p.controllerCallsign).Logger()
	if p.enableTextLogging {
		logger = logger.With().Str("text", tx).Logger()
	}
	logger.Debug().Msg("parsing text")
	tx = normalize(tx)
	if tx == "" {
		return nil
	}
	tx = p.uncrushCallsign(tx)

	if p.enableTextLogging {
		logger = logger.With().Str("normalized", tx).Logger()
	}
	logger.Debug().Msg("normalized text")

	fields := strings.Fields(tx)

	before := fields
	var requestArgs []string
	requestWord, idx, ok := findRequestWord(fields)
	if ok {
		logger = logger.With().Str("request", requestWord).Logger()
		logger.Debug().Int("position", idx).Msg("found request word")
		before, requestArgs = fields[:idx], fields[idx+1:]
	}

	heard, rest, ok := p.findControllerCallsign(before)

	if !ok {
		logger.Trace().Msg("no GCI callsign found")
		return nil
	}
	event := logger.Debug().Str("heard", heard)
	if p.enableTextLogging {
		event = event.Str("rest", rest)
	}
	event.Msg("found GCI callsign")

	event = logger.Debug()
	if p.enableTextLogging {
		event = event.Str("rest", rest)
	}
	event.Msg("searching for pilot callsign in rest of text")

	rest = numwords.ParseString(rest)
	pilotCallsign, ok := ParsePilotCallsign(rest)
	if ok {
		logger = logger.With().Str("pilot", pilotCallsign).Logger()
		logger.Debug().Msg("found pilot callsign")
	} else {
		logger.Trace().Msg("no pilot callsign found")
	}

	if !ok {
		if requestWord != "" && requestWord == picture {
			return &brevity.PictureRequest{Callsign: ""}
		}
		return &brevity.UnableToUnderstandRequest{}
	}
	if requestWord == "" {
		logger.Trace().Msg("no request word found")
		return handleNoRequestWord(tx, pilotCallsign)
	}

	switch requestWord {
	case alphaCheck:
		return &brevity.AlphaCheckRequest{Callsign: pilotCallsign}
	case radioCheck:
		return &brevity.RadioCheckRequest{Callsign: pilotCallsign}
	case picture:
		return &brevity.PictureRequest{Callsign: pilotCallsign}
	case tripwire:
		return &brevity.TripwireRequest{Callsign: pilotCallsign}
	case shopping:
		return &brevity.ShoppingRequest{Callsign: pilotCallsign}
	}

	event = logger.Debug()
	if p.enableTextLogging {
		event = event.Strs("args", requestArgs)
	}
	event.Msg("parsing request arguments")

	return parseRequestWithArgs(requestWord, pilotCallsign, requestArgs)
}
