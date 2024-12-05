package application

import (
	"context"
	"errors"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// recognize runs speech recognition on audio received from SRS and forwards recognized text to the given channel.
func (a *Application) recognize(ctx context.Context, out chan<- Message[string]) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping speech recognition due to context cancellation")
			return
		case transmission := <-a.srsClient.Receive():
			rCtx := context.Background()
			rCtx = traces.WithTraceID(rCtx, transmission.TraceID)
			rCtx = traces.WithClientName(rCtx, transmission.ClientName)
			rCtx = traces.WithReceivedAt(rCtx, time.Now())
			a.recognizeSample(ctx, rCtx, transmission.Audio, out)
		}
	}
}

// recognizeSample runs speech recognition on a single audio sample and forwards the recognized text to the output channel.
// The first context is the parent context of the process, and the second context is the context of the request.
// If the recognition process takes longer than 30 seconds, recognizeSample will log an error and return without publishing a message.
func (a *Application) recognizeSample(processCtx context.Context, requestCtx context.Context, audio simpleradio.Audio, out chan<- Message[string]) {
	recogizerCtx, cancel := context.WithTimeout(processCtx, 30*time.Second)
	defer func() {
		if recogizerCtx.Err() != nil && errors.Is(recogizerCtx.Err(), context.DeadlineExceeded) {
			a.trace(traces.WithRequestError(requestCtx, recogizerCtx.Err()))
		}
	}()
	defer cancel()

	log.Info().Msg("recognizing audio sample")
	start := time.Now()
	text, err := a.recognizer.Recognize(recogizerCtx, audio, a.enableTranscriptionLogging)
	if err != nil {
		log.Error().Err(err).Msg("error recognizing audio sample")
		a.trace(traces.WithRequestError(processCtx, err))
		return
	}
	logger := log.With().Stringer("clockTime", time.Since(start)).Logger()

	requestCtx = traces.WithRecognizedAt(requestCtx, time.Now())
	requestCtx = traces.WithRequestText(requestCtx, text)
	if a.enableTranscriptionLogging {
		logger = logger.With().Str("text", text).Logger()
	}
	logger.Info().Msg("recognized audio")
	out <- AsMessage(requestCtx, text)
}
