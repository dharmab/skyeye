// parser converts converts brevity requests from natural language into structured forms.
package parser

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/brevity"
	fuzz "github.com/hbollon/go-edlib"
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
	"ready":     radioCheck,
	"read your": radioCheck,
	"bogeido":   bogeyDope,
	"bokeido":   bogeyDope,
	"bokey":     bogeyDope,
	"bokeh":     bogeyDope,
	"bogeydope": bogeyDope,
	"okey":      bogeyDope,
	"boogie":    bogeyDope,
	"oogie":     bogeyDope,
	"snap lock": snaplock,
}

func IsSimilar(a, b string) bool {
	v, err := fuzz.StringsSimilarity(a, b, fuzz.Levenshtein)
	if err != nil {
		log.Error().Err(err).Str("a", a).Str("b", b).Msg("failed to calculate similarity")
		return false
	}
	return v > 0.7
}

func (p *parser) findGCICallsign(fields []string) (string, int, bool) {
	found := false
	candidate := ""
	pivot := -1
	for i, field := range fields {
		candidate += field
		for _, wakePhrase := range []string{p.gciCallsign, Anyface} {
			if IsSimilar(candidate, strings.ToLower(wakePhrase)) {
				found = true
				pivot = i
				break
			}
		}
	}
	return candidate, pivot, found
}

func findRequestWord(fields []string) (string, int, bool) {
	for i, field := range fields {
		for _, word := range requestWords {
			if IsSimilar(string(word), field) {
				return word, i, true
			}
		}
	}
	return "", 0, false
}

// Parse implements Parser.Parse.
func (p *parser) Parse(tx string) any {
	logger := log.With().Str("gci", p.gciCallsign).Logger()
	logger.Debug().Str("text", tx).Msg("parsing text")
	tx = strings.ToLower(tx)
	tx = strings.ReplaceAll(tx, ",", "")
	tx = strings.ReplaceAll(tx, ".", "")
	tx = strings.ReplaceAll(tx, "-", " ")
	tx = strings.TrimSpace(tx)
	for alt, word := range alternateRequestWords {
		tx = strings.ReplaceAll(tx, alt, string(word))
	}
	logger = logger.With().Str("text", tx).Logger()
	logger.Debug().Msg("normalized text")

	if tx == "" {
		return nil
	}

	// Tokenize the text.
	fields := strings.Fields(tx)

	// Search for a token that looks similar to a request word, and split
	// the text around it.
	before := fields
	var requestArgs []string
	requestWord, requestWordIndex, foundRequestWord := findRequestWord(fields)
	if foundRequestWord {
		logger = logger.With().Str("request", string(requestWord)).Logger()
		logger.Debug().Int("position", requestWordIndex).Msg("found request word")
		before, requestArgs = fields[:requestWordIndex], fields[requestWordIndex+1:]
	}

	// Search the first part of the text for text that looks similar to a GCI
	// callsign. If we find such text, search the rest for a valid pilot
	// callsign.
	_, pivot, foundGCICallsign := p.findGCICallsign(before)
	if foundGCICallsign {
		logger.Debug().Int("pivot", pivot).Msg("found GCI callsign")
	}

	// If we didn't hear the GCI callsign, this was probably chatter rather
	// than a request.
	if !foundGCICallsign {
		logger.Trace().Msg("no GCI callsign found")
		return nil
	}

	if len(before) < pivot {
		logger.Trace().Msg("nothing left to search for pilot callsign")
		return &brevity.UnableToUnderstandRequest{}
	}

	pilotCallsign, foundPilotCallsign := ParsePilotCallsign(strings.Join(before[pivot+1:], " "))
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
	case picture:
		if request, ok := p.parsePicture(pilotCallsign, scanner); ok {
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
	return &brevity.UnableToUnderstandRequest{Callsign: pilotCallsign}
}

var numberWords = map[string]int{
	"0":    0,
	"zero": 0,
	//"o":     0,
	"oh":    0,
	"1":     1,
	"one":   1,
	"wun":   1,
	"2":     2,
	"two":   2,
	"3":     3,
	"three": 3,
	"tree":  3,
	"4":     4,
	"four":  4,
	"fower": 4,
	"5":     5,
	"five":  5,
	"fife":  5,
	"6":     6,
	"six":   6,
	"7":     7,
	"seven": 7,
	"8":     8,
	"eight": 8,
	"ait":   8,
	"9":     9,
	"nine":  9,
	"niner": 9,
}

// ParsePilotCallsign attempts to parse a callsign in one of the following formats:
//
// - A single word, followed by a number consisting of any digits
//
// - A number consisting of any digits
//
// Garbage in between the digits is ignored. The result is normalized so that each digit is lowercase and space-delimited.
func ParsePilotCallsign(tx string) (callsign string, isValid bool) {
	tx, _, _ = strings.Cut(tx, "|")
	tx, _, _ = strings.Cut(tx, "#")
	tx = strings.Trim(tx, " ")
	for i, char := range tx {
		if unicode.IsDigit(char) {
			tx = fmt.Sprintf("%s %s", strings.TrimSpace(tx[:i]), strings.TrimSpace(tx[i:]))
			break
		}
	}
	var scanner = bufio.NewScanner(strings.NewReader(tx))
	scanner.Split(bufio.ScanWords)

	ok := scanner.Scan()
	if !ok {
		return
	}
	firstToken := scanner.Text()
	if firstToken == "" {
		return
	}
	callsign, ok = appendNumber(callsign, firstToken)
	if !ok {
		callsign = firstToken
	} else {
		isValid = true
	}

	for scanner.Scan() {
		nextToken := scanner.Text()
		// Handle single digit
		s, ok := appendNumber(callsign, nextToken)
		if ok {
			callsign = s
			isValid = true
		} else {
			// Handle case where multiple digits are not space-delimited
			for _, char := range nextToken {
				s, ok := appendNumber(callsign, string(char))
				if ok {
					callsign = s
					isValid = true
				}
			}
			if !isValid {
				callsign = fmt.Sprintf("%s%s", callsign, nextToken)
			}
		}
	}
	callsign = strings.ToLower(callsign)
	return
}

func appendNumber(callsign string, number string) (string, bool) {
	if d, ok := numberWords[number]; ok {
		return fmt.Sprintf("%s %d", callsign, d), true
	}
	return callsign, false
}
