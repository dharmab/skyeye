package parser

import "github.com/dharmab/skyeye/pkg/brevity"

type alphaCheckRequest struct {
	callsign string
}

var _ brevity.AlphaCheckRequest = &alphaCheckRequest{}

func (r *alphaCheckRequest) Callsign() string {
	return r.callsign
}
