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
	gciCallsign string
}

func New(callsign string) Parser {
	return &parser{
		gciCallsign: strings.ReplaceAll(callsign, " ", ""),
	}
}

const Anyface string = "anyface"

const (
	alphaCheck string = "alpha"
	bogeyDope  string = "bogey"
	declare    string = "declare"
	picture    string = "picture"
	radioCheck string = "radio"
	spiked     string = "spiked"
	snaplock   string = "snaplock"
)

var requestWords = []string{radioCheck, alphaCheck, bogeyDope, declare, picture, spiked, snaplock}

var alternateRequestWords = map[string]string{
	"radiocheck": radioCheck,
	"bogeido":    bogeyDope,
	"bokeido":    bogeyDope,
	"bokeydope":  bogeyDope,
	"bokey":      bogeyDope,
	"bokeh":      bogeyDope,
	"bogy":       bogeyDope,
	"bogeydope":  bogeyDope,
	"okey":       bogeyDope,
	"boogie":     bogeyDope,
	"oogie":      bogeyDope,
	"foggydope":  bogeyDope,
	"snap lock":  snaplock,
}

func IsSimilar(a, b string) bool {
	v, err := fuzz.StringsSimilarity(strings.ToLower(a), strings.ToLower(b), fuzz.Levenshtein)
	if err != nil {
		log.Error().Err(err).Str("a", a).Str("b", b).Msg("failed to calculate similarity")
		return false
	}
	return v > 0.49
}

func (p *parser) findGCICallsign(fields []string) (string, string, bool) {
	for i := range fields {
		candidate := strings.Join(fields[:i+1], " ")
		for _, wakePhrase := range []string{p.gciCallsign, Anyface} {
			if IsSimilar(strings.TrimSpace(candidate), strings.ToLower(wakePhrase)) {
				return candidate, strings.Join(fields[i+1:], " "), true
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

func normalize(tx string) string {
	tx, _, _ = strings.Cut(tx, "|")
	tx = strings.ToLower(tx)
	tx = strings.ReplaceAll(tx, "-", " ")
	for _, r := range tx {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			tx = strings.ReplaceAll(tx, string(r), "")
		}
	}
	tx = strings.TrimSpace(tx)
	for alt, word := range alternateRequestWords {
		tx = strings.ReplaceAll(tx, alt, word)
	}
	tx = strings.Join(strings.Fields(tx), " ")
	return tx
}

func spaceDigits(tx string) string {
	txBuilder := strings.Builder{}
	for _, char := range tx {
		if unicode.IsDigit(char) {
			txBuilder.WriteRune(' ')
		}
		txBuilder.WriteRune(char)
	}
	tx = txBuilder.String()
	return normalize(tx)
}

// Parse implements Parser.Parse.
func (p *parser) Parse(tx string) any {
	logger := log.With().Str("gci", p.gciCallsign).Logger()
	logger.Debug().Str("text", tx).Msg("parsing text")
	tx = normalize(tx)
	if tx == "" {
		return nil
	}
	logger = logger.With().Str("text", tx).Logger()
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
	} else {
		logger.Debug().Str("heard", heardGCICallsign).Str("after", afterGCICallsign).Msg("found GCI callsign")
	}

	logger.Debug().Str("rest", afterGCICallsign).Msg("searching for pilot callsign in rest of text")
	afterGCICallsign = numwords.ParseString(afterGCICallsign)
	pilotCallsign, foundPilotCallsign := ParsePilotCallsign(afterGCICallsign)
	if foundPilotCallsign {
		logger = logger.With().Str("pilot", pilotCallsign).Logger()
		logger.Debug().Msg("found pilot callsign")
	}

	// Handle cases where we heard our own callsign, but couldn't understand
	// the request.
	if !foundPilotCallsign {
		logger.Trace().Msg("no pilot callsign found")
		return &brevity.UnableToUnderstandRequest{}
	}
	if !foundRequestWord {
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
	}

	logger.Debug().Strs("args", requestArgs).Msg("parsing request arguments")
	scanner := bufio.NewScanner(strings.NewReader(strings.Join(requestArgs, " ")))
	scanner.Split(bufio.ScanWords)

	switch requestWord {
	case bogeyDope:
		if request, ok := p.parseBogeyDope(pilotCallsign, scanner); ok {
			return request
		}
	case declare:
		if request, ok := p.parseDeclare(pilotCallsign, scanner); ok {
			return request
		}
	case spiked:
		if request, ok := p.parseSpiked(pilotCallsign, scanner); ok {
			return request
		}
	case snaplock:
		if request, ok := p.parseSnaplock(pilotCallsign, scanner); ok {
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
			builder.WriteRune(char)
		}
	}

	callsign = spaceDigits(normalize(builder.String()))
	if callsign == "" {
		return "", false
	}

	return callsign, true

}
