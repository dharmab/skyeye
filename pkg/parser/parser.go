package parser

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rodaine/numwords"
	"github.com/rs/zerolog/log"
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

func New(callsign string) Parser {
	return &parser{
		callsign: callsign,
	}
}

const anyface string = "anyface"

var anyfaceWords = []string{
	"any face",
	"any phase",
}

type requestWord string

const (
	alphaCheck requestWord = "alpha"
	bogeyDope  requestWord = "bogey"
	declare    requestWord = "declare"
	picture    requestWord = "picture"
	radioCheck requestWord = "radio"
	spike      requestWord = "spike"
	spiked     requestWord = "spiked"
	snaplock   requestWord = "snaplock"
)

var alternateRequestWords = map[string]requestWord{
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

func requestWords() []requestWord {
	return []requestWord{alphaCheck, bogeyDope, declare, picture, radioCheck, spiked, spike, snaplock}
}

func (p *parser) parseWakeWord(scanner *bufio.Scanner) (string, bool) {
	ok := scanner.Scan()
	if !ok {
		return "", false
	}
	firstSegment := scanner.Text()
	if !(firstSegment == strings.ToLower(p.callsign) || firstSegment == anyface) {
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
		log.Info().Str("callsign", p.callsign).Str("text", tx).Msg("no wake word found in text")
		return nil, false
	}
	log.Debug().Str("callsign", p.callsign).Str("text", tx).Msg("found wake word")

	// Scan until we find a request trigger word. Split the scanned tranmission into a callsign segment and a request word.
	var segment string
	callsign := ""
	var rWord requestWord
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for callsign == "" {
		select {
		case <-ctx.Done():
			log.Warn().Str("text", tx).Msg("timed out parsing callsign")
			return &brevity.UnableToUnderstandRequest{}, true
		default:
			ok := scanner.Scan()
			if !ok {
				log.Debug().Str("text", tx).Msg("no request word found in text")
				return &brevity.UnableToUnderstandRequest{}, true
			}

			segment = fmt.Sprintf("%s %s", segment, scanner.Text())

			for k, v := range alternateRequestWords {
				if strings.Contains(segment, k) {
					log.Debug().Str("segment", segment).Str("alternate", k).Str("canonical", string(v)).Msg("replacing request word")
					segment = strings.Replace(segment, k, string(v), 1)
					break
				}
			}

			for _, word := range requestWords() {
				select {
				case <-ctx.Done():
					log.Warn().Str("text", tx).Msg("timed out parsing callsign and request")
					return &brevity.UnableToUnderstandRequest{Callsign: callsign}, true

				default:
					if strings.HasSuffix(strings.TrimSpace(segment), string(word)) {
						log.Debug().Str("segment", segment).Str("request word", string(word)).Msg("found request word")
						rWord = word
						log.Debug().Str("segment", segment).Msg("parsing callsign")
						callsignSegment := strings.TrimSuffix(segment, string(word))
						callsignSegment = p.sanitize(callsignSegment)
						callsign, ok = ParseCallsign(callsignSegment)
						if !ok {
							log.Debug().Str("segment", segment).Msg("unable to parse request callsign")
							return &brevity.UnableToUnderstandRequest{
								Callsign: "",
							}, true
						}
						log.Debug().Str("callsign", callsign).Str("segment", segment).Msg("parsed callsign")
						if len(callsign) > 30 {
							log.Warn().Str("callsign", callsign).Msg("callsign too long, ignoring request")
							return nil, false
						}
						_ = scanner.Scan()
						log.Debug().Str("text", tx).Str("request", string(word)).Msg("found request word")
						break
					}
				}
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
	case spike:
		return p.parseSpiked(callsign, scanner)
	case spiked:
		return p.parseSpiked(callsign, scanner)
	case snaplock:
		return p.parseSnaplock(callsign, scanner)
	default:
		return &brevity.UnableToUnderstandRequest{}, false
	}
}

var sanitizerRex = regexp.MustCompile(`[^\p{L}\p{N} ]+`)

// sanitize lowercases the input and replaces punctuation with spaces.
func (p *parser) sanitize(s string) string {
	log.Debug().Str("text", s).Msg("sanitizing text")
	lowercased := strings.ToLower(s)
	if s != lowercased {
		log.Debug().Str("text", lowercased).Msg("lowercased text")
	}
	numbersCleaned := numwords.ParseString(lowercased)
	if lowercased != numbersCleaned {
		log.Debug().Str("text", numbersCleaned).Msg("parsed numbers")
	}
	punctuationCleaned := sanitizerRex.ReplaceAllString(numbersCleaned, " ")
	if numbersCleaned != punctuationCleaned {
		log.Debug().Str("text", punctuationCleaned).Msg("replaced punctuation")
	}
	finalClean := punctuationCleaned
	for _, words := range anyfaceWords {
		if strings.HasPrefix(finalClean, words) {
			finalClean = strings.Replace(finalClean, words, "anyface", 1)
			log.Debug().Str("text", finalClean).Msg("cleaned up ANYFACE")
			break
		}
	}
	log.Debug().Str("text", finalClean).Msg("sanitized text")
	return finalClean
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
// Garbage in between the digits is ignored. The result is normalized so that each digit is lowercase and space-delimited.
func ParseCallsign(tx string) (callsign string, isValid bool) {
	tx = strings.Trim(tx, " ")
	for i, char := range tx {
		if unicode.IsDigit(char) {
			newTx := fmt.Sprintf("%s %s", tx[:i], tx[i:])
			log.Trace().Str("text", tx).Str("separated", newTx).Msg("separating letters and digits")
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
	callsign = strings.ToLower(callsign)
	return
}

func appendNumber(callsign string, number string) (string, bool) {
	if d, ok := numberWords[number]; ok {
		return fmt.Sprintf("%s %d", callsign, d), true
	}
	return callsign, false
}
