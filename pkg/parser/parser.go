// Package parser converts brevity requests from natural language into structured forms.
package parser

import (
	"bufio"
	"bytes"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
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
	tripwire   string = "tripwire"
)

var requestWords = []string{radioCheck, alphaCheck, bogeyDope, declare, picture, spiked, snaplock, tripwire, shopping}

func (p *Parser) findControllerCallsign(fields []string) (callsign string, rest string, found bool) {
	for i := range fields {
		candidate := strings.Join(fields[:i+1], " ")
		for _, wakePhrase := range []string{p.controllerCallsign, Anyface} {
			if isSimilar(strings.TrimSpace(candidate), strings.ToLower(wakePhrase)) {
				found = true
				callsign = candidate
				rest = strings.Join(fields[i+1:], " ")
				return
			}
		}
	}
	return "", "", false
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
			if len(field) > 8 {
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

	// Tokenize the text.
	fields := strings.Fields(tx)

	// Search for a token that looks similar to a request word, and split
	// the text around it.
	before := fields
	var requestArgs []string
	requestWord, requestWordIndex, foundRequestWord := findRequestWord(fields)
	if foundRequestWord {
		logger = logger.With().Str("request", requestWord).Logger()
		logger.Debug().Int("position", requestWordIndex).Msg("found request word")
		before, requestArgs = fields[:requestWordIndex], fields[requestWordIndex+1:]
	}

	// Search the first part of the text for text that looks similar to a GCI
	// callsign. If we find such text, search the rest for a valid pilot
	// callsign.
	heardControllerCallsign, afterControllerCallsign, foundControllerCallsign := p.findControllerCallsign(before)

	// If we didn't hear the GCI callsign, this was probably chatter rather
	// than a request.
	if !foundControllerCallsign {
		logger.Trace().Msg("no GCI callsign found")
		return nil
	}
	event := logger.Debug().Str("heard", heardControllerCallsign)
	if p.enableTextLogging {
		event = event.Str("rest", afterControllerCallsign)
	}
	event.Msg("found GCI callsign")
	logger.Debug().Str("heard", heardControllerCallsign).Str("after", afterControllerCallsign).Msg("found GCI callsign")

	event = logger.Debug()
	if p.enableTextLogging {
		event = event.Str("rest", afterControllerCallsign)
	}
	event.Msg("searching for pilot callsign in rest of text")

	afterControllerCallsign = numwords.ParseString(afterControllerCallsign)
	pilotCallsign, foundPilotCallsign := ParsePilotCallsign(afterControllerCallsign)
	if foundPilotCallsign {
		logger = logger.With().Str("pilot", pilotCallsign).Logger()
		logger.Debug().Msg("found pilot callsign")
	}

	// Handle cases where we heard our own callsign, but couldn't understand
	// the request.
	if !foundPilotCallsign && foundRequestWord && requestWord == picture {
		return &brevity.PictureRequest{Callsign: ""}
	}
	if !foundPilotCallsign {
		logger.Trace().Msg("no pilot callsign found")
		return &brevity.UnableToUnderstandRequest{}
	}
	if !foundRequestWord {
		// Fallback: Possibly an ambiguous check-in request.
		if strings.Contains(tx, checkIn) {
			return &brevity.CheckInRequest{Callsign: pilotCallsign}
		}

		logger.Trace().Msg("no request word found")
		return &brevity.UnableToUnderstandRequest{Callsign: pilotCallsign}
	}

	// Try to parse a request from the remaining text.
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
	scanner := bufio.NewScanner(strings.NewReader(strings.Join(requestArgs, " ")))
	scanner.Split(bufio.ScanWords)

	switch requestWord {
	case bogeyDope:
		if request, ok := parseBogeyDope(pilotCallsign, scanner); ok {
			return request
		}
	case declare:
		if request, ok := parseDeclare(pilotCallsign, scanner); ok {
			return request
		}
	case spiked:
		if request, ok := parseSpiked(pilotCallsign, scanner); ok {
			return request
		}
	case snaplock:
		if request, ok := parseSnaplock(pilotCallsign, scanner); ok {
			return request
		}
	}

	logger.Debug().Msg("unrecognized request")
	return &brevity.UnableToUnderstandRequest{Callsign: pilotCallsign}
}

func skipWords(scanner *bufio.Scanner, words ...string) bool {
	for _, word := range words {
		if isSimilar(scanner.Text(), word) {
			return scanner.Scan()
		}
	}
	return true
}

func prependToScanner(scanner *bufio.Scanner, s string) *bufio.Scanner {
	if s == "" {
		return scanner
	}
	var buffer bytes.Buffer
	_, _ = buffer.WriteString(s + " ")

	for scanner.Scan() {
		_, _ = buffer.WriteString(scanner.Text() + " ")
	}
	newScanner := bufio.NewScanner(&buffer)
	newScanner.Split(bufio.ScanWords)
	return newScanner
}
