package controller

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/jba/omap"
	"github.com/rs/zerolog/log"
)

// findCallsign uses fuzzy matching to find a trackfile for the given callsign.
// Any matching callsign is returned, along with any trackfile and a bool indicating
// if a valid trackfile with a non-zero location was found.
func (c *Controller) findCallsign(callsign string) (string, *trackfiles.Trackfile, bool) {
	logger := log.With().Str("parsedCallsign", callsign).Logger()
	foundCallsign, trackfile := c.scope.FindCallsign(callsign, c.coalition)
	if trackfile == nil {
		logger.Info().Msg("no trackfile found for callsign")
		return "", nil, false
	}
	logger = logger.With().Str("foundCallsign", foundCallsign).Logger()
	if trackfile.IsLastKnownPointZero() {
		logger.Info().Msg("found trackfile for callsign but without location")
		return foundCallsign, trackfile, false
	}
	logger.Debug().Msg("found trackfile for callsign")
	return foundCallsign, trackfile, true
}

// strictCallsign is a strucured model of a callsign that includes the flight
// codeword (which may include spaces) and a two-digit flight and position
// number.
type strictCallsign struct {
	codeword string
	flight   int
	position int
}

func (c *strictCallsign) String() string {
	return fmt.Sprintf("%s %d %d", c.codeword, c.flight, c.position)
}

// parseStrictCallsign parses a freeform callsign into a strict callsign, if possible.
func parseStrictCallsign(s string) (*strictCallsign, bool) {
	words := strings.Fields(s)
	if len(words) < 3 {
		return nil, false
	}

	flight, err := strconv.Atoi(words[len(words)-2])
	if err != nil {
		return nil, false
	}
	if flight < 0 {
		return nil, false
	}

	element, err := strconv.Atoi(words[len(words)-1])
	if err != nil {
		return nil, false
	}
	if element < 0 {
		return nil, false
	}

	codeword := strings.Join(words[:len(words)-2], " ")
	if codeword == "" {
		return nil, false
	}
	for _, r := range codeword {
		if unicode.IsDigit(r) {
			return nil, false
		}
	}

	return &strictCallsign{
		codeword: codeword,
		flight:   flight,
		position: element,
	}, true
}

// getFriendlyCallsigns returns all parseable callsigns for all friendly trackfiles.
func (c *Controller) getFriendlyCallsigns() []string {
	friendlies := c.scope.FindByCoalition(c.coalition)
	callsigns := []string{}
	for _, friendly := range friendlies {
		if callsign, ok := parser.ParsePilotCallsign(friendly.Contact.Name); ok {
			callsigns = append(callsigns, callsign)
		}
	}
	return callsigns
}

// collateCallsigns combines callsigns into flights and elements, as documented in SkyEye's player guide.
//
// receivers is a slice of callsigns that should be collated.
//
// everyone is a slice of all friendly callsigns, whether they are receivers or not.
//
// Returns a slice of callsigns which are a best-effort collation of the receivers.
func collateCallsigns(receivers, everyone []string) []string {
	// Short-circuit if there is only one receiver.
	if len(receivers) <= 1 {
		return receivers
	}

	// Sort the slices for a consistent order.
	slices.Sort(receivers)
	slices.Sort(everyone)

	// Parse all callsigns which follow the strict format.
	callsigns := []strictCallsign{}
	for _, name := range everyone {
		if callsign, ok := parseStrictCallsign(name); ok {
			callsigns = append(callsigns, *callsign)
		}
	}

	// Sort the strict callsigns by codeword, flight, and position. This
	// organizes the callsigns in the final broadcast call.
	slices.SortFunc(callsigns, func(a, b strictCallsign) int {
		if a.codeword != b.codeword {
			return strings.Compare(a.codeword, b.codeword)
		}
		if a.flight != b.flight {
			return a.flight - b.flight
		}
		return a.position - b.position
	})

	// Build an ordered trie of callsigns, where the first level is the
	// codeword, the second level is the flight, and the third level is the
	// position. The value is a boolean indicating whether the callsign is a
	// receiver.
	members := omap.Map[string, *omap.Map[int, *omap.Map[int, bool]]]{}
	for _, callsign := range callsigns {
		flights, ok := members.Get(callsign.codeword)
		if !ok {
			flights = &omap.Map[int, *omap.Map[int, bool]]{}
			members.Set(callsign.codeword, flights)
		}
		positions, ok := flights.Get(callsign.flight)
		if !ok {
			positions = &omap.Map[int, bool]{}
			flights.Set(callsign.flight, positions)
		}
		isReceiver := slices.Contains(receivers, callsign.String())
		positions.Set(callsign.position, isReceiver)
	}

	// Walk the trie. If an entire subtrie is receivers, then the entire flight
	// is a collatable receiver. Otherwise, iterate over the positions in the
	// flight and collate the receivers by flight and position. Collect the
	// collated callsigns in a slice.
	collated := []string{}
	for codeword, flights := range members.All() {
		for flight, positions := range flights.All() {
			// Note: Iterating twice over the positions. Probaby fine since flights are only 4-5 positions at most.
			isEntireFlightReceiver := true
			for _, isReceiver := range positions.All() {
				if !isReceiver {
					isEntireFlightReceiver = false
					break
				}
			}
			if isEntireFlightReceiver && positions.Len() > 1 {
				collated = append(collated, fmt.Sprintf("%s %d flight", codeword, flight))
			} else {
				isFirstReceiverInFlight := true
				for position, isReceiver := range positions.All() {
					if !isReceiver {
						continue
					}
					var formatted string
					if isFirstReceiverInFlight {
						formatted = fmt.Sprintf("%s %d %d", codeword, flight, position)
						isFirstReceiverInFlight = false
					} else {
						formatted = fmt.Sprintf("%d %d", flight, position)
					}
					collated = append(collated, formatted)
				}
			}
		}
	}

	// Append any non-strict callsigns to the collated callsigns.
	for _, s := range receivers {
		if _, ok := parseStrictCallsign(s); !ok {
			collated = append(collated, s)
		}
	}

	return collated
}
