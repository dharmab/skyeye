// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer"
)

// Application is the interface for running the SkyEye application.
type Application interface {
	// Run runs the SkyEye application. It should be called exactly once.
	Run(context.Context) error
}

// app implements the Application.
type app struct {
	// dcsClient is a DCS-gRPC client
	dcsClient dcs.DCSClient
	// srsClient is a SimpleRadio Standalone client
	srsClient simpleradio.Client
	// recognizer provides speech-to-text recognition
	recognizer recognizer.Recognizer
	// parser converts English brevity text to internal representations
	parser parser.Parser
	// controller implements internal GCI logic
	controller controller.Controller
	// composer converys from internal representations to English brevity text
	composer composer.Composer
	// synthesizer provides text-to-speech synthesis
	synthesizer synthesizer.Sythesizer
}

// NewApplication constructs a new Application.
func NewApplication(ctx context.Context, config conf.Configuration) (Application, error) {
	slog.Info("constructing DCS client")
	dcsClient, err := dcs.NewDCSClient(
		ctx,
		dcs.ClientConfiguration{
			Address:           config.DCSGRPCAddress,
			ConnectionTimeout: config.GRPCConnectionTimeout,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructing SRS client")
	srsClient, err := simpleradio.NewClient(
		srs.ClientConfiguration{
			Address:                   config.SRSAddress,
			ConnectionTimeout:         config.SRSConnectionTimeout,
			ClientName:                config.SRSClientName,
			ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
			Coalition:                 config.SRSCoalition,
			Radio: srs.Radio{
				Frequency:        config.SRSFrequency,
				Modulation:       srs.ModulationAM,
				ShouldRetransmit: true,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructing speech-to-text recognizer")
	recognizer := recognizer.NewWhisperRecognizer(config.WhisperModel)

	slog.Info("constructing text parser")
	parser := parser.New()

	slog.Info("constructing GCI controller")
	controller := controller.New()

	slog.Info("constructing text composer")
	composer := composer.New()

	slog.Info("constructing text-to-speech synthesizer")
	synthesizer, err := synthesizer.NewPiperSpeaker(synthesizer.FeminineVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructiong application")
	app := &app{
		dcsClient:   dcsClient,
		srsClient:   srsClient,
		recognizer:  recognizer,
		parser:      parser,
		controller:  controller,
		composer:    composer,
		synthesizer: synthesizer,
	}
	return app, nil
}

// Run implements Application.Run.
func (a *app) Run(ctx context.Context) error {
	defer func() {
		slog.Info("closing connection to DCS-gRPC server")
		err := a.dcsClient.Close()
		if err != nil {
			slog.Error("failed to close connection to DCS-gRPC server", "error", err)
		}
	}()

	go func() {
		slog.Info("running SRS client")
		if err := a.srsClient.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("error running SRS client", "error", err)
			}
		}
	}()

	rxTextChan := make(chan string)
	requestChan := make(chan any)
	responseAndCallsChan := make(chan any)
	txTextChan := make(chan composer.NaturalLanguageResponse)
	txAudioChan := make(chan []float32)

	go a.recognize(ctx, rxTextChan)
	go a.parse(ctx, rxTextChan, requestChan)
	go a.control(ctx, requestChan, responseAndCallsChan)
	go a.compose(ctx, responseAndCallsChan, txTextChan)
	go a.synthesize(ctx, txTextChan, txAudioChan)
	go a.transmit(ctx, txAudioChan)

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping application due to context cancellation")
			return nil
		case sample := <-a.srsClient.Receive():
			slog.Info("recognizing audio sample received from SRS client")
			text, err := a.recognizer.Recognize(sample)
			if err != nil {
				slog.Error("error recongizing audio sample", "error", err)
			} else {
				slog.Info("recognized audio", "text", text)
				rxTextChan <- text
			}
		}
	}
}

// recognize runs speech recognition on audio received from SRS and forwards recognized text to the given channel.
func (a *app) recognize(ctx context.Context, out chan<- string) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping speech recognition due to context cancellation")
			return
		case sample := <-a.srsClient.Receive():
			slog.Info("recognizing audio sample")
			text, err := a.recognizer.Recognize(sample)
			if err != nil {
				slog.Error("error recongizing audio sample", "error", err)
			} else if text == "" {
				slog.Debug("unable to recongnize any words in audio sample")
			} else {
				slog.Info("recognized audio", "text", text)
				out <- text
			}
		}
	}
}

// parse converts incoming brevity from text format to internal representations.
func (a *app) parse(ctx context.Context, in <-chan string, out chan<- any) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping brevity parsing due to context cancellation")
			return
		case text := <-in:
			slog.Info("parsing text", "text", text)
			request, ok := a.parser.Parse(text)
			if ok {
				slog.Info("parsed text", "group", request)
				out <- request
			} else {
				slog.Debug("unable to parse text", "text", text)
			}
		}
	}
}

// control routes requests to GCI controller handlers.
func (a *app) control(ctx context.Context, in <-chan any, out chan<- any) {
	slog.Info("running controller")
	go a.controller.Run(ctx, out)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping request routing due to context cancellation")
			return
		case brev := <-in:
			slog.Info("routing request to controller", "request", brev)
			switch request := brev.(type) {
			case brevity.AlphaCheckRequest:
				slog.Info("routing ALPHA CHECK request to controller", "request", request)
				a.controller.HandleAlphaCheck(request)
			case brevity.BogeyDopeRequest:
				slog.Info("routing BOGEY DOPE request to controller", "request", request)
				a.controller.HandleBogeyDope(request)
			case brevity.DeclareRequest:
				slog.Info("routing DECLARE request to controller", "request", request)
				a.controller.HandleDeclare(request)
			case brevity.PictureRequest:
				slog.Info("routing PICTURE request to controller", "request", request)
				a.controller.HandlePicture(request)
			case brevity.RadioCheckRequest:
				slog.Info("routing RADIO CHECK request to controller", "request", request)
				a.controller.HandleRadioCheck(request)
			case brevity.SnaplockRequest:
				slog.Info("routing SNAPLOCK request to controller", "request", request)
				a.controller.HandleSnaplock(request)
			case brevity.SpikedRequest:
				slog.Info("routing SPIKED request to controller", "request", request)
				a.controller.HandleSpiked(request)
			default:
				slog.Error("unable to route request to handler", "request", request)
			}
		}
	}
}

// compose converts outgoing brevity from internal representations to text format.
func (a *app) compose(ctx context.Context, in <-chan any, out chan<- composer.NaturalLanguageResponse) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping brevity composition due to context cancellation")
			return
		case call := <-in:
			slog.Info("composing brevity call", "call", call)
			var nlr composer.NaturalLanguageResponse
			switch c := call.(type) {
			case brevity.AlphaCheckResponse:
				slog.Info("composing ALPHA CHECK call", "call", c)
				nlr = a.composer.ComposeAlphaCheckResponse(c)
			case brevity.BogeyDopeResponse:
				slog.Info("composing BOGEY DOPE call", "call", c)
				nlr = a.composer.ComposeBogeyDopeResponse(c)
			case brevity.DeclareResponse:
				slog.Info("composing DECLARE call", "call", c)
				nlr = a.composer.ComposeDeclareResponse(c)
			case brevity.FadedCall:
				slog.Info("composing FADED call", "call", c)
				nlr = a.composer.ComposeFadedCall(c)
			case brevity.PictureResponse:
				slog.Info("composing PICTURE call", "call", c)
				nlr = a.composer.ComposePictureResponse(c)
			case brevity.RadioCheckResponse:
				slog.Info("composing RADIO CHECK call", "call", c)
				nlr = a.composer.ComposeRadioCheckResponse(c)
			case brevity.SnaplockResponse:
				slog.Info("composing SNAPLOCK call", "call", c)
				nlr = a.composer.ComposeSnaplockResponse(c)
			case brevity.SpikedResponse:
				slog.Info("composing SPIKED call", "call", c)
				nlr = a.composer.ComposeSpikedResponse(c)
			case brevity.SunriseCall:
				slog.Info("composing SUNRISE call", "call", c)
				nlr = a.composer.ComposeSunriseCall(c)
			case brevity.ThreatCall:
				slog.Info("composing THREAT call", "call", c)
				nlr = a.composer.ComposeThreatCall(c)
			default:
				slog.Error("unable to route call to composition", "call", call)
			}
			if nlr.Speech != "" && nlr.Subtitle != "" {
				slog.Info("composed brevity call", "speech", nlr.Speech, "subtitle", nlr.Subtitle)
				out <- nlr
			}
		}
	}
}

// synthesize converts outgoing text to spoken audio.
func (a *app) synthesize(ctx context.Context, in <-chan composer.NaturalLanguageResponse, out chan<- []float32) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping speech synthesis due to context cancellation")
			return
		case response := <-in:
			slog.Info("synthesizing speech", "text", response.Speech)
			audio, err := a.synthesizer.Say(response.Speech)
			if err != nil {
				slog.Error("error synthesizing speech", "error", err)
			} else {
				slog.Info("synthesized speech", "text", response.Speech)
				out <- audio
			}
		}
	}
}

// transmit sends audio to SRS for transmission.
func (a *app) transmit(ctx context.Context, in <-chan []float32) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping audio transmissions due to context cancellation")
			return
		case audio := <-in:
			slog.Info("transmitting audio")
			a.srsClient.Transmit(audio)
		}
	}
}
