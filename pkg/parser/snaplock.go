package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func parseSnaplock(callsign string, scanner *bufio.Scanner) (*brevity.SnaplockRequest, bool) {
	bra, ok := parseBRA(scanner)
	if !ok {
		return nil, false
	}
	return &brevity.SnaplockRequest{
		Callsign: callsign,
		BRA:      bra,
	}, true
}
