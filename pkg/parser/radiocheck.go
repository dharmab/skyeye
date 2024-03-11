package parser

import "github.com/dharmab/skyeye/pkg/brevity"

type radioCheckRequest struct {
	callsign string
}

var _ brevity.RadioCheckRequest = &radioCheckRequest{}

func (r *radioCheckRequest) Callsign() string {
	return r.callsign
}
