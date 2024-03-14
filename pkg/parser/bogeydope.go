package parser

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

type bogeyDopeRequest struct {
	callsign string
	filter   brevity.BogeyFilter
}

var _ brevity.BogeyDopeRequest = &bogeyDopeRequest{}

func (r *bogeyDopeRequest) BogeyDope() {}

func (r *bogeyDopeRequest) Callsign() string {
	return r.callsign
}

func (r *bogeyDopeRequest) Filter() brevity.BogeyFilter {
	return r.filter
}

var bogeyFilterMap = map[string]brevity.BogeyFilter{
	"airplane":    brevity.Airplanes,
	"planes":      brevity.Airplanes,
	"fighter":     brevity.Airplanes,
	"fixed wing":  brevity.Airplanes,
	"helicopter":  brevity.Helicopters,
	"chopper":     brevity.Helicopters,
	"helo":        brevity.Helicopters,
	"rotary wing": brevity.Helicopters,
}

func (p *parser) parseBogeyDope(callsign string, scanner *bufio.Scanner) (brevity.BogeyDopeRequest, bool) {
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
	return &bogeyDopeRequest{callsign, filter}, true
}
