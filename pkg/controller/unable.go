package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

func (c *controller) HandleUnableToUnderstand(ctx context.Context, request *brevity.UnableToUnderstandRequest) {
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	response := brevity.SayAgainResponse{Callsign: "last caller"}
	if callsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition); trackfile != nil {
		response.Callsign = callsign
	}
	c.calls <- NewCall(ctx, response)
}
