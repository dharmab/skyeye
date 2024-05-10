package parser

import (
	"bufio"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rodaine/numwords"
)

type Parser interface {
	// Parse reads a transmission from a player and returns an intermediate representation (IR) of the transmitted request.
	// The IR's type must be determined by reflection. The boolean return value is true if the transmission was parsed into a valid IR
	// and false otherwise. If the return value is false, the IR must be nil.
	Parse(string) (any, bool)
}

type parser struct {
	// callsign of the GCI
	callsign string
}

func New() Parser {
	return &parser{
		callsign: "skyeye",
	}
}

const anyface string = "anyface"

var anyfaceWords = []string{
	"any face",
	"any phase",
}

type requestWord string

const (
	alphaCheck requestWord = "alpha check"
	bogeyDope  requestWord = "bogey dope"
	declare    requestWord = "declare"
	picture    requestWord = "picture"
	radioCheck requestWord = "radio check"
	spiked     requestWord = "spiked"
	snaplock   requestWord = "snaplock"
)

var alternateRequestWords = map[string]requestWord{
	"ready 1 check":   radioCheck,
	"read your check": radioCheck,
	"radio chat":      radioCheck,
	"radio jack":      radioCheck,
}

func requestWords() []requestWord {
	return []requestWord{alphaCheck, bogeyDope, declare, picture, radioCheck, spiked, snaplock}
}

func (p *parser) parseWakeWord(scanner *bufio.Scanner) (string, bool) {
	ok := scanner.Scan()
	if !ok {
		return "", false
	}
	firstSegment := scanner.Text()
	if !(firstSegment == p.callsign || firstSegment == anyface) {
		return "", false
	}
	return firstSegment, true
}

// Parse implements Parser.Parse.
func (p *parser) Parse(tx string) (any, bool) {
	tx = p.sanitize(tx)

	scanner := bufio.NewScanner(strings.NewReader(tx))
	scanner.Split(bufio.ScanWords)

	// Check for a wake word (GCI callsign)
	_, ok := p.parseWakeWord(scanner)
	if !ok {
		slog.Info("no wake word found in text", "text", tx)
		return nil, false
	}

	// Scan until we find a request trigger word. Split the scanned tranmission into a callsign segment and a request word.
	var segment string
	callsign := ""
	var rWord requestWord
	for callsign == "" {
		ok := scanner.Scan()
		if !ok {
			return nil, false
		}

		segment = fmt.Sprintf("%s %s", segment, scanner.Text())

		for k, v := range alternateRequestWords {
			if strings.Contains(segment, k) {
				segment = strings.Replace(segment, k, string(v), 1)
				break
			}
		}

		for _, word := range requestWords() {
			if strings.HasSuffix(segment, string(word)) {
				rWord = word
				// Try to parse a callsign from the second segment.
				callsignSegment := strings.TrimSuffix(segment, string(word))
				callsignSegment = p.sanitize(callsignSegment)
				callsign, ok = ParseCallsign(callsignSegment)
				if !ok {
					// TODO send "say again" response?
					return nil, false
				}
				if len(callsign) > 30 {
					return nil, false
				}
				_ = scanner.Scan()

				break
			}
		}
	}

	// Try to parse a request from the remaining text in the scanner.
	switch rWord {
	case alphaCheck:
		// ALPHA CHECK, as implemented by this bot, is a simple request.
		return &brevity.AlphaCheckRequest{Callsign: callsign}, true
	case bogeyDope:
		return p.parseBogeyDope(callsign, scanner)
	case declare:
		return p.parseDeclare(callsign, scanner)
	case picture:
		return p.parsePicture(callsign, scanner)
	case radioCheck:
		// RADIO CHECK is a simple request.
		return &brevity.RadioCheckRequest{Callsign: callsign}, true
	case spiked:
		return p.parseSpiked(callsign, scanner)
	case snaplock:
		return p.parseSnaplock(callsign, scanner)
	}
	return nil, false
}

var sanitizerRex = regexp.MustCompile(`[^\p{L}\p{N} ]+`)

// sanitize lowercases the input and replaces punctuation with spaces.
func (p *parser) sanitize(s string) string {
	s = strings.ToLower(s)
	s = numwords.ParseString(s)
	s = sanitizerRex.ReplaceAllString(s, " ")
	for _, words := range anyfaceWords {
		if strings.HasPrefix(s, words) {
			s = strings.Replace(s, words, "anyface", 1)
			break
		}
	}
	slog.Info("sanitized text", "text", s)
	return s
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

// ParseCallsign attempts to parse a callsign in one of the following formats:
//
// - A single word, followed by a number consisting of any digits
//
// - A number consisting of any digits
//
// Garbage in between the digits is ignored. The result is normalized so that each digit is space-delimited.
func ParseCallsign(tx string) (callsign string, isValid bool) {
	tx = strings.Trim(tx, " ")
	for i, char := range tx {
		if unicode.IsDigit(char) {
			newTx := fmt.Sprintf("%s %s", tx[:i], tx[i:])
			slog.Info("separating letters and digits", "text", tx, "separated", newTx)
			tx = newTx
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
	return
}

func appendNumber(callsign string, number string) (string, bool) {
	if d, ok := numberWords[number]; ok {
		return fmt.Sprintf("%s %d", callsign, d), true
	}
	return callsign, false
}
