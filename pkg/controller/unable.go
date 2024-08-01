package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *controller) HandleUnableToUnderstand(request *brevity.UnableToUnderstandRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	c.out <- brevity.SayAgainResponse{Callsign: request.Callsign}
}
