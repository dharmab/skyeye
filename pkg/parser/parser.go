// parser converts converts brevity requests from natural language into structured forms.
package parser

import (
	"bufio"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/brevity"
	fuzz "github.com/hbollon/go-edlib"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
)

type Parser interface {
	// Parse reads natural language text, checks if it starts with the GCI
	// callsign, and attempts to parse a request from the text. Returns a
	// brevity request, or nil if the text does not start with the GCI
	// callsign.
	Parse(string) any
}

type parser struct {
	gciCallsign       string
	enableTextLogging bool
}

func New(callsign string, enableTextLogging bool) Parser {
	return &parser{
		gciCallsign:       strings.ReplaceAll(callsign, " ", ""),
		enableTextLogging: enableTextLogging,
	}
}

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

func IsSimilar(a, b string) bool {
	v, err := fuzz.StringsSimilarity(strings.ToLower(a), strings.ToLower(b), fuzz.Levenshtein)
	if err != nil {
		log.Error().Err(err).Str("a", a).Str("b", b).Msg("failed to calculate similarity")
		return false
	}
	return v > 0.6
}

func (p *parser) findGCICallsign(fields []string) (callsign string, rest string, found bool) {
	for i := range fields {
		candidate := strings.Join(fields[:i+1], " ")
		for _, wakePhrase := range []string{p.gciCallsign, Anyface} {
			if IsSimilar(strings.TrimSpace(candidate), strings.ToLower(wakePhrase)) {
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
		for _, word := range requestWords {
			if IsSimilar(word, field) {
				return word, i, true
			}
		}
	}
	return "", 0, false
}

// normalize the given string by applying the following transformations:
//
//   - Split on any "|" character and discard the tail.
//   - Convert to lowercase.
//   - Replace hyphens and underscores with spaces. Remove any other characters
//     that are not letters, digits, or spaces.
//   - Insert a space between any letter immediately followed by a digit.
//   - Trim leading and trailing whitespace.
//   - Substitute alternate forms of request words with canonical forms.
//   - Remove extra spaces.
func normalize(tx string) string {
	tx, _, _ = strings.Cut(tx, "|")
	tx = strings.ToLower(tx)
	tx = removeSymbols(tx)
	tx = spaceNumbers(tx)
	tx = strings.TrimSpace(tx)
	for alt, word := range alternateRequestWords {
		tx = strings.ReplaceAll(tx, alt, word)
	}
	tx = strings.Join(strings.Fields(tx), " ")
	return tx
}

// removeSymbols removes any characters that are not letters, digits, or spaces.
// Hyphens and underscores are replaced with spaces. Other symbols are removed.
func removeSymbols(tx string) string {
	var builder strings.Builder
	for _, r := range tx {
		if r == '-' || r == '_' {
			_, _ = builder.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			_, _ = builder.WriteRune(r)
		}
	}
	return builder.String()
}

// spaceNumbers inserts spaces between letters and numbers, e.g. "BRAA090" -> "BRAA 090".
func spaceNumbers(tx string) string {
	builder := strings.Builder{}
	for i, char := range tx {
		_, _ = builder.WriteRune(char)
		if i+1 < len(tx) && unicode.IsLetter(char) && unicode.IsDigit(rune(tx[i+1])) {
			_, _ = builder.WriteRune(' ')
		}
	}
	return builder.String()
}

// spaceDigits inserts a space before each digit, e.g. "Eagle11" -> "Eagle 1 1".
func spaceDigits(tx string) string {
	builder := strings.Builder{}
	for _, char := range tx {
		if unicode.IsDigit(char) {
			_, _ = builder.WriteRune(' ')
		}
		_, _ = builder.WriteRune(char)
	}
	tx = builder.String()
	return normalize(tx)
}

// uncrushCallsign corrects a corner case where the GCI callsign and the
// following token have no space between them, e.g. "anyfaceeagle 1".
func (p *parser) uncrushCallsign(s string) string {
	for _, callsign := range []string{p.gciCallsign, Anyface} {
		lc := strings.ToLower(callsign)
		if strings.HasPrefix(s, lc) {
			return lc + " " + s[len(lc):]
		}
	}
	return s
}

// Parse implements Parser.Parse.
func (p *parser) Parse(tx string) any {
	logger := log.With().Str("gci", p.gciCallsign).Logger()
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
	heardGCICallsign, afterGCICallsign, foundGCICallsign := p.findGCICallsign(before)

	// If we didn't hear the GCI callsign, this was probably chatter rather
	// than a request.
	if !foundGCICallsign {
		logger.Trace().Msg("no GCI callsign found")
		return nil
	}
	event := logger.Debug().Str("heard", heardGCICallsign)
	if p.enableTextLogging {
		event = event.Str("rest", afterGCICallsign)
	}
	event.Msg("found GCI callsign")
	logger.Debug().Str("heard", heardGCICallsign).Str("after", afterGCICallsign).Msg("found GCI callsign")

	event = logger.Debug()
	if p.enableTextLogging {
		event = event.Str("rest", afterGCICallsign)
	}
	event.Msg("searching for pilot callsign in rest of text")

	afterGCICallsign = numwords.ParseString(afterGCICallsign)
	pilotCallsign, foundPilotCallsign := ParsePilotCallsign(afterGCICallsign)
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

// ParsePilotCallsign attempts to parse a callsign in one of the following formats:
//   - A single word, followed by a number consisting of any digits
//   - A number consisting of up to 3 digits
//
// Garbage in between the digits is ignored. The result is normalized so that each digit is lowercase and space-delimited.
func ParsePilotCallsign(tx string) (callsign string, isValid bool) {
	tx = normalize(tx)
	tx = spaceDigits(tx)
	tx = strings.ReplaceAll(tx, "request", "")
	tx = strings.ReplaceAll(tx, "this is", "")

	var builder strings.Builder
	numDigits := 0
	for _, char := range tx {
		if numDigits >= 3 {
			break
		}
		if unicode.IsDigit(char) {
			numDigits++
		}
		if numDigits == 0 || unicode.IsDigit(char) || unicode.IsSpace(char) {
			_, _ = builder.WriteRune(char)
		}
	}

	callsign = spaceDigits(normalize(builder.String()))
	if callsign == "" {
		return "", false
	}

	return callsign, true
}

func skipWords(scanner *bufio.Scanner, words ...string) bool {
	for _, word := range words {
		if IsSimilar(scanner.Text(), word) {
			return scanner.Scan()
		}
	}
	return true
}
