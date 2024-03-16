package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
)

func (p *parser) parsePicture(callsign string, scanner *bufio.Scanner) (*brevity.PictureRequest, bool) {
	radius, ok := p.parseRange(scanner)
	if !ok {
		radius = conf.DefaultPictureRadius
	}
	return &brevity.PictureRequest{
		Callsign: callsign,
		Radius:   radius,
	}, true
}
