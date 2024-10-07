package application

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// synthesize converts outgoing text to spoken audio.
func (a *app) synthesize(ctx context.Context, in <-chan Message[composer.NaturalLanguageResponse], out chan<- Message[simpleradio.Audio]) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping speech synthesis due to context cancellation")
			return
		case message := <-in:
			a.synthesizeMessage(message.Context, message.Data, out)
		}
	}
}

// synthesizeMessage synthesizes a single message and publishes the audio to the output channel.
func (a *app) synthesizeMessage(ctx context.Context, response composer.NaturalLanguageResponse, out chan<- Message[simpleradio.Audio]) {
	log.Info().Str("text", response.Speech).Msg("synthesizing speech")
	start := time.Now()
	audio, err := a.speaker.Say(response.Speech)
	if err != nil {
		log.Error().Err(err).Msg("error synthesizing speech")
		a.trace(traces.WithRequestError(ctx, err))
	} else {
		log.Info().Stringer("clockTime", time.Since(start)).Msg("synthesized audio")
		out <- AsMessage(
			traces.WithSynthesizedAt(ctx, time.Now()),
			simpleradio.Audio(audio),
		)
	}
}
