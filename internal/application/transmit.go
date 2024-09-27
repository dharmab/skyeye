package application

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// transmit sends audio to SRS for transmission.
func (a *app) transmit(ctx context.Context, in <-chan Message[simpleradio.Audio]) {
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
func (a *app) transmitMessage(rCtx context.Context, audio simpleradio.Audio) {
	log.Info().Msg("transmitting audio")
	transmission := simpleradio.Transmission{
		TraceID:    traces.GetTraceID(rCtx),
		ClientName: traces.GetClientName(rCtx),
		Audio:      audio,
	}
	a.srsClient.Transmit(transmission)
	a.trace(traces.WithSubmittedAt(rCtx, time.Now()))
}
