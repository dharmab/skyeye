package application

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// transmit sends audio to SRS for transmission.
func (a *Application) transmit(ctx context.Context, in <-chan Message[simpleradio.Audio]) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping audio transmissions due to context cancellation")
			return
		case message := <-in:
			a.transmitMessage(message.Context, message.Data)
		}
	}
}

// transmitMessage submits a single audio sample to SRS.
func (a *Application) transmitMessage(rCtx context.Context, audio simpleradio.Audio) {
	transmission := simpleradio.Transmission{
		TraceID:    traces.GetTraceID(rCtx),
		ClientName: traces.GetClientName(rCtx),
		Audio:      audio,
	}

	log.Info().Str("traceID", transmission.TraceID).Msg("transmitting audio")
	a.srsClient.Transmit(transmission)
	a.trace(traces.WithSubmittedAt(rCtx, time.Now()))
}
