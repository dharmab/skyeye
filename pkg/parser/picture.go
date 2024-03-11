package parser

import (
	"bufio"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

type pictureRequest struct {
	callsign string
	radius   unit.Length
}

var _ brevity.PictureRequest = &pictureRequest{}

func (r *pictureRequest) Callsign() string {
	return r.callsign
}

func (r *pictureRequest) Radius() unit.Length {
	return r.radius
}

func (p *parser) parsePicture(callsign string, scanner *bufio.Scanner) (brevity.PictureRequest, bool) {
	radius, ok := p.parseRange(scanner)
	if !ok {
		radius = conf.DefaultPictureRadius
	}
	return &pictureRequest{callsign, radius}, true
}
