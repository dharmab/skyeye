// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer"
	"github.com/rs/zerolog/log"
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
	log.Info().Str("address", config.DCSGRPCAddress).Dur("timeout", config.GRPCConnectionTimeout).Msg("constructing DCS client")
	dcsClient, err := dcs.NewDCSClient(
		ctx,
		dcs.ClientConfiguration{
			Address:           config.DCSGRPCAddress,
			ConnectionTimeout: config.GRPCConnectionTimeout,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	log.Info().
		Str("address", config.SRSAddress).
		Dur("timeout", config.SRSConnectionTimeout).
		Str("clientName", config.SRSClientName).
		Int("coalitionID", int(config.SRSCoalition)).
		Float64("frequency", config.SRSFrequency).
		Int("modulationID", int(srs.ModulationAM)).
		Msg("constructing SRS client")
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

	log.Info().Msg("constructing speech-to-text recognizer")
	recognizer := recognizer.NewWhisperRecognizer(config.WhisperModel)

	log.Info().Msg("constructing text parser")
	parser := parser.New()

	log.Info().Msg("constructing radar scope")
	var coalition types.Coalition
	if config.SRSCoalition == srs.CoalitionBlue {
		coalition = types.CoalitionBlue
	} else {
		coalition = types.CoalitionRed
	}

	updates := make(chan dcs.Updated)
	fades := make(chan dcs.Faded)
	bullseyes := make(chan dcs.Bullseye)

	rdr := radar.New(coalition, bullseyes, updates, fades)
	log.Info().Msg("constructing GCI controller")
	controller := controller.New(rdr, coalition)

	log.Info().Msg("constructing text composer")
	composer := composer.New()

	log.Info().Msg("constructing text-to-speech synthesizer")
	synthesizer, err := synthesizer.NewPiperSpeaker(synthesizer.FeminineVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	log.Info().Msg("constructing application")
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
		log.Info().Msg("closing connection to DCS-gRPC server")
		err := a.dcsClient.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close connection to DCS-gRPC server")
		}
	}()

	go func() {
		log.Info().Msg("running SRS client")
		if err := a.srsClient.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("error running SRS client")
			}
		}
	}()

	rxTextChan := make(chan string)
	requestChan := make(chan any)
	responseAndCallsChan := make(chan any)
	txTextChan := make(chan composer.NaturalLanguageResponse)
	txAudioChan := make(chan []float32)

	log.Info().Msg("starting subroutines")
	log.Info().Msg("starting speech recognition routine")
	go a.recognize(ctx, rxTextChan)
	log.Info().Msg("starting speech-to-text parsing routine")
	go a.parse(ctx, rxTextChan, requestChan)
	log.Info().Msg("starting GCI controller routine")
	go a.control(ctx, requestChan, responseAndCallsChan)
	log.Info().Msg("starting response composer routine")
	go a.compose(ctx, responseAndCallsChan, txTextChan)
	log.Info().Msg("starting speech synthesis routine")
	go a.synthesize(ctx, txTextChan, txAudioChan)
	log.Info().Msg("starting radio transmission routine")
	go a.transmit(ctx, txAudioChan)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping application due to context cancellation")
			return nil
		case sample := <-a.srsClient.Receive():
			log.Info().Int("byteLength", len(sample)).Msg("recognizing audio sample received from SRS client")
			text, err := a.recognizer.Recognize(sample)
			if err != nil {
				log.Error().Err(err).Msg("error recognizing audio sample")
			} else {
				// TODO make this log line configurable for privacy reasons
				log.Info().Str("text", text).Msg("recognized audio sample")
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
			log.Info().Msg("stopping speech recognition due to context cancellation")
			return
		case sample := <-a.srsClient.Receive():
			log.Info().Int("byteLength", len(sample)).Msg("recognizing audio sample")
			text, err := a.recognizer.Recognize(sample)
			if err != nil {
				log.Error().Err(err).Msg("error recognizing audio sample")
			} else if text == "" || text == "[BLANK AUDIO]\n" {
				log.Info().Str("text", text).Msg("unable to recongnize any words in audio sample")
			} else {
				log.Info().Str("text", text).Msg("recognized audio")
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
			log.Info().Msg("stopping text parsing due to context cancellation")
			return
		case text := <-in:
			logger := log.With().Str("text", text).Logger()
			logger.Info().Msg("parsing text")
			request, ok := a.parser.Parse(text)
			if ok {
				logger.Info().Interface("request", request).Msg("parsed text")
				out <- request
			} else {
				logger.Info().Msg("unable to parse text")
			}
		}
	}
}

// control routes requests to GCI controller handlers.
func (a *app) control(ctx context.Context, in <-chan any, out chan<- any) {
	log.Info().Msg("running controller")
	go a.controller.Run(ctx, out)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping controller request routing due to context cancellation")
			return
		case brev := <-in:
			logger := log.With().Type("type", brev).Logger()
			logger.Info().Msg("routing request to controller")
			switch request := brev.(type) {
			case *brevity.AlphaCheckRequest:
				logger.Debug().Msg("routing ALPHA CHECK request to controller")
				a.controller.HandleAlphaCheck(request)
			case *brevity.BogeyDopeRequest:
				logger.Debug().Msg("routing BOGEY DOPE request to controller")
				a.controller.HandleBogeyDope(request)
			case *brevity.DeclareRequest:
				logger.Debug().Msg("routing DECLARE request to controller")
				a.controller.HandleDeclare(request)
			case *brevity.PictureRequest:
				logger.Debug().Msg("routing PICTURE request to controller")
				a.controller.HandlePicture(request)
			case *brevity.RadioCheckRequest:
				logger.Debug().Msg("routing RADIO CHECK request to controller")
				a.controller.HandleRadioCheck(request)
			case *brevity.SnaplockRequest:
				logger.Debug().Msg("routing SNAPLOCK request to controller")
				a.controller.HandleSnaplock(request)
			case *brevity.SpikedRequest:
				logger.Debug().Msg("routing SPIKED request to controller")
				a.controller.HandleSpiked(request)
			default:
				logger.Error().Interface("request", brev).Msg("unable to route request to handler")
			}
		}
	}
}

// compose converts outgoing brevity from internal representations to text format.
func (a *app) compose(ctx context.Context, in <-chan any, out chan<- composer.NaturalLanguageResponse) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping brevity composition due to context cancellation")
			return
		case call := <-in:
			logger := log.With().Interface("call", call).Logger()
			logger.Info().Msg("composing brevity call")
			var nlr composer.NaturalLanguageResponse
			switch c := call.(type) {
			case brevity.AlphaCheckResponse:
				logger.Debug().Msg("composing ALPHA CHECK call")
				nlr = a.composer.ComposeAlphaCheckResponse(c)
			case brevity.BogeyDopeResponse:
				logger.Debug().Msg("composing BOGEY DOPE call")
				nlr = a.composer.ComposeBogeyDopeResponse(c)
			case brevity.DeclareResponse:
				logger.Debug().Msg("composing DECLARE call")
				nlr = a.composer.ComposeDeclareResponse(c)
			case brevity.FadedCall:
				logger.Debug().Msg("composing FADED call")
				nlr = a.composer.ComposeFadedCall(c)
			case brevity.PictureResponse:
				logger.Debug().Msg("composing PICTURE call")
				nlr = a.composer.ComposePictureResponse(c)
			case brevity.RadioCheckResponse:
				logger.Debug().Msg("composing RADIO CHECK call")
				nlr = a.composer.ComposeRadioCheckResponse(c)
			case brevity.SnaplockResponse:
				logger.Debug().Msg("composing SNAPLOCK call")
				nlr = a.composer.ComposeSnaplockResponse(c)
			case brevity.SpikedResponse:
				logger.Debug().Msg("composing SPIKED call")
				nlr = a.composer.ComposeSpikedResponse(c)
			case brevity.SunriseCall:
				logger.Debug().Msg("composing SUNRISE call")
				nlr = a.composer.ComposeSunriseCall(c)
			case brevity.ThreatCall:
				logger.Debug().Msg("composing THREAT call")
				nlr = a.composer.ComposeThreatCall(c)
			default:
				logger.Debug().Msg("unable to route call to composition")
			}
			if nlr.Speech != "" && nlr.Subtitle != "" {
				logger.Info().Str("speech", nlr.Speech).Str("subtitle", nlr.Subtitle).Msg("composed brevity call")
				out <- nlr
			} else {
				logger.Warn().Msg("natural language response is empty")
			}
		}
	}
}

// synthesize converts outgoing text to spoken audio.
func (a *app) synthesize(ctx context.Context, in <-chan composer.NaturalLanguageResponse, out chan<- []float32) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping speech synthesis due to context cancellation")
			return
		case response := <-in:
			log.Info().Str("text", response.Speech).Msg("synthesizing speech")
			audio, err := a.synthesizer.Say(response.Speech)
			if err != nil {
				log.Error().Err(err).Msg("error synthesizing speech")
			} else {
				if len(audio) == 0 {
					log.Warn().Msg("synthesized audio is empty")
				} else {
					log.Info().Int("byteLength", len(audio)).Msg("synthesized audio")
					out <- audio
				}
			}
		}
	}
}

// transmit sends audio to SRS for transmission.
func (a *app) transmit(ctx context.Context, in <-chan []float32) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping audio transmissions due to context cancellation")
			return
		case audio := <-in:
			if len(audio) == 0 {
				log.Warn().Msg("audio to transmit is empty")
			} else {
				log.Info().Int("byteLength", len(audio)).Msg("transmitting audio")
			}
			a.srsClient.Transmit(audio)
		}
	}
}
