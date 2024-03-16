package parser

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

var bogeyFilterMap = map[string]brevity.ContactCategory{
	"airplane":    brevity.Airplanes,
	"planes":      brevity.Airplanes,
	"fighter":     brevity.Airplanes,
	"fixed wing":  brevity.Airplanes,
	"helicopter":  brevity.Helicopters,
	"chopper":     brevity.Helicopters,
	"helo":        brevity.Helicopters,
	"rotary wing": brevity.Helicopters,
}

func (p *parser) parseBogeyDope(callsign string, scanner *bufio.Scanner) (*brevity.BogeyDopeRequest, bool) {
	filter := brevity.Everything
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
