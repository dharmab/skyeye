package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// HandleUnableToUnderstand handles requests where the wake word was recognized but the request could not be understood, by asking players on the channel to repeat their message.
func (c *Controller) HandleUnableToUnderstand(ctx context.Context, request *brevity.UnableToUnderstandRequest) {
	unableCounter.Add(ctx, 1)
	log.Debug().Str("callsign", request.Callsign).Type("type", request).Msg("handling request")
	response := brevity.SayAgainResponse{Callsign: brevity.LastCaller}
	if callsign, trackfile := c.scope.FindCallsign(request.Callsign, c.coalition); trackfile != nil {
		response.Callsign = callsign
	}
	c.calls <- NewCall(ctx, response)
}
