package application

import (
	"context"
	"errors"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// compose converts outgoing brevity from internal representations to text format.
func (a *Application) compose(ctx context.Context, in <-chan controller.Call, out chan<- Message[composer.NaturalLanguageResponse]) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping brevity composition due to context cancellation")
			return
		case call := <-in:
			a.composeCall(call.Context, call.Call, out)
		}
	}
}

// composeCall handles a single call, publishing the composition to the output channel.
func (a *Application) composeCall(ctx context.Context, call any, out chan<- Message[composer.NaturalLanguageResponse]) {
	ctx = traces.WithHandledAt(ctx, time.Now())
	logger := log.With().Type("type", call).Any("params", call).Logger()
	logger.Info().Msg("composing brevity call")
	var response composer.NaturalLanguageResponse
	switch c := call.(type) {
	case brevity.AlphaCheckResponse:
		response = a.composer.ComposeAlphaCheckResponse(c)
	case brevity.BogeyDopeResponse:
		response = a.composer.ComposeBogeyDopeResponse(c)
	case brevity.CheckInResponse:
		response = a.composer.ComposeCheckInResponse(c)
	case brevity.DeclareResponse:
		response = a.composer.ComposeDeclareResponse(c)
	case brevity.FadedCall:
		response = a.composer.ComposeFadedCall(c)
	case brevity.NegativeRadarContactResponse:
		response = a.composer.ComposeNegativeRadarContactResponse(c)
	case brevity.PictureResponse:
		response = a.composer.ComposePictureResponse(c)
	case brevity.RadioCheckResponse:
		response = a.composer.ComposeRadioCheckResponse(c)
	case brevity.ShoppingResponse:
		response = a.composer.ComposeShoppingResponse(c)
	case brevity.SnaplockResponse:
		response = a.composer.ComposeSnaplockResponse(c)
	case brevity.SpikedResponseV2:
		response = a.composer.ComposeSpikedResponse(c)
	case brevity.StrobeResponse:
		response = a.composer.ComposeStrobeResponse(c)
	case brevity.TripwireResponse:
		response = a.composer.ComposeTripwireResponse(c)
	case brevity.SunriseCall:
		response = a.composer.ComposeSunriseCall(c)
	case brevity.ThreatCall:
		response = a.composer.ComposeThreatCall(c)
	case brevity.MergedCall:
		response = a.composer.ComposeMergedCall(c)
	case brevity.SayAgainResponse:
		response = a.composer.ComposeSayAgainResponse(c)
	default:
		logger.Debug().Msg("unable to route call to composition")
		a.trace(traces.WithRequestError(ctx, errors.New("no route for call")))
	}

	logger.Info().Str("speech", response.Speech).Str("subtitle", response.Subtitle).Msg("composed brevity call")
	ctx = traces.WithCallText(ctx, response.Subtitle)
	ctx = traces.WithComposedAt(ctx, time.Now())
	out <- AsMessage(ctx, response)
}
