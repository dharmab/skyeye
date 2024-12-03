package parser

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

var bogeyFilterMap = map[string]brevity.ContactCategory{
	"airplane":    brevity.FixedWing,
	"planes":      brevity.FixedWing,
	"fighter":     brevity.FixedWing,
	"fixed wing":  brevity.FixedWing,
	"helicopter":  brevity.RotaryWing,
	"chopper":     brevity.RotaryWing,
	"helo":        brevity.RotaryWing,
	"rotary wing": brevity.RotaryWing,
}

func parseBogeyDope(callsign string, scanner *bufio.Scanner) (*brevity.BogeyDopeRequest, bool) {
	filter := brevity.Aircraft
	s := scanner.Text()
	for scanner.Scan() {
		s = fmt.Sprintf("%s %s", s, scanner.Text())
	}
	for k, v := range bogeyFilterMap {
		if strings.Contains(s, k) {
			filter = v
			break
		}
	}
	return &brevity.BogeyDopeRequest{Callsign: callsign, Filter: filter}, true
}
