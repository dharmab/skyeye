package application

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// parse converts incoming brevity from text format to internal representations.
func (a *app) parse(ctx context.Context, in <-chan Message[string], out chan<- Message[any]) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping text parsing due to context cancellation")
			return
		case message := <-in:
			a.parseText(message.Context, message.Data, out)
		}
	}
}

// parseText parses a single transcribed transmission, publishing any successfully parsed requests to the output channel.
func (a *app) parseText(ctx context.Context, text string, out chan<- Message[any]) {
	logger := log.Logger
	if a.enableTranscriptionLogging {
		logger = logger.With().Str("text", text).Logger()
	}
	logger.Info().Msg("parsing text")
	request := a.parser.Parse(text)
	ctx = traces.WithParsedAt(ctx, time.Now())
	if request != nil {
		ctx = traces.WithRequest(ctx, request)
		logger.Info().Any("request", request).Msg("parsed text")
		out <- AsMessage(ctx, request)
	} else {
		logger.Info().Msg("unable to parse text, could be silence, chatter, missing GCI callsign")
		a.trace(ctx)
	}
}
