package application

import (
	"context"
	"errors"
	"sync"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// control runs the GCI controller and routes requests to the controller.
func (a *Application) control(ctx context.Context, wg *sync.WaitGroup, in <-chan Message[any], out chan<- controller.Call) {
	log.Info().Msg("running controller")
	wg.Go(func() {
		a.controller.Run(ctx, out)
	})
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping controller request routing due to context cancellation")
			return
		case message := <-in:
			a.handleRequest(message.Context, message.Data)
		}
	}
}

// handleRequest routes the given request to the controller's appropriate handler.
func (a *Application) handleRequest(ctx context.Context, r any) {
	logger := log.With().Type("type", a).Logger()
	logger.Info().Msg("routing request to controller")
	switch request := r.(type) {
	case *brevity.AlphaCheckRequest:
		a.controller.HandleAlphaCheck(ctx, request)
	case *brevity.BogeyDopeRequest:
		a.controller.HandleBogeyDope(ctx, request)
	case *brevity.CheckInRequest:
		a.controller.HandleCheckIn(ctx, request)
	case *brevity.DeclareRequest:
		a.controller.HandleDeclare(ctx, request)
	case *brevity.PictureRequest:
		a.controller.HandlePicture(ctx, request)
	case *brevity.RadioCheckRequest:
		a.controller.HandleRadioCheck(ctx, request)
	case *brevity.ShoppingRequest:
		a.controller.HandleShopping(ctx, request)
	case *brevity.SnaplockRequest:
		a.controller.HandleSnaplock(ctx, request)
	case *brevity.SpikedRequest:
		a.controller.HandleSpiked(ctx, request)
	case *brevity.TripwireRequest:
		a.controller.HandleTripwire(ctx, request)
	case *brevity.UnableToUnderstandRequest:
		a.controller.HandleUnableToUnderstand(ctx, request)
	default:
		logger.Error().Any("request", request).Msg("unable to route request to handler")
		a.trace(traces.WithRequestError(ctx, errors.New("no route for request")))
	}
}
