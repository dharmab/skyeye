package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *controller) HandleUnableToUnderstand(r *brevity.UnableToUnderstandRequest) {
	log.Debug().Str("callsign", r.Callsign).Type("type", r).Msg("handling request")
	c.out <- brevity.SayAgainResponse{Callsign: r.Callsign}
}
