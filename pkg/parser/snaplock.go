package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

type snaplockRequest struct {
	callsign string
	bra      brevity.BRA
}

var _ brevity.SnaplockRequest = &snaplockRequest{}

func (r *snaplockRequest) Callsign() string {
	return r.callsign
}

func (r *snaplockRequest) BRA() brevity.BRA {
	return r.bra
}

func (p *parser) parseSnaplock(callsign string, scanner *bufio.Scanner) (brevity.SnaplockRequest, bool) {
	bra, ok := p.parseBRA(scanner)
	if !ok {
		return nil, false
	}
	return &snaplockRequest{
		callsign: callsign,
		bra:      bra,
	}, true
}
