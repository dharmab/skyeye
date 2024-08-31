// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/simpleradio/audio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	tacview "github.com/dharmab/skyeye/pkg/tacview/client"
	"github.com/rs/zerolog/log"
)

// Application is the interface for running the SkyEye application.
type Application interface {
	// Run runs the SkyEye application. It should be called exactly once.
	Run(context.Context, context.CancelFunc, *sync.WaitGroup) error
}

// app implements the Application.
type app struct {
	// srsClient is a SimpleRadio Standalone client
	srsClient     simpleradio.Client
	tacviewClient tacview.Client
	// recognizer provides speech-to-text recognition
	recognizer recognizer.Recognizer
	// parser converts English brevity text to internal representations
	parser parser.Parser
	radar  radar.Radar
	// controller implements internal GCI logic
	controller controller.Controller
	// composer converys from internal representations to English brevity text
	composer composer.Composer
	// speaker provides text-to-speech synthesis
	speaker speakers.Speaker
}

// NewApplication constructs a new Application.
func NewApplication(ctx context.Context, config conf.Configuration) (Application, error) {
	updates := make(chan sim.Updated)
	fades := make(chan sim.Faded)

	log.Info().
		Str("address", config.SRSAddress).
		Stringer("timeout", config.SRSConnectionTimeout).
		Str("clientName", config.SRSClientName).
		Int("coalitionID", int(config.Coalition)).
		Float64("frequency", config.SRSFrequency.Megahertz()).
		Int("modulationID", int(srs.ModulationAM)).
		Msg("constructing SRS client")
	srsClient, err := simpleradio.NewClient(
		srs.ClientConfiguration{
			Address:                   config.SRSAddress,
			ConnectionTimeout:         config.SRSConnectionTimeout,
			ClientName:                config.SRSClientName,
			ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
			Coalition:                 config.Coalition,
			Radio: srs.Radio{
				Frequency:        config.SRSFrequency.Hertz(),
				Modulation:       srs.ModulationAM,
				ShouldRetransmit: true,
			},
			Mute: config.Mute,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	var tacviewClient tacview.Client
	if config.ACMIFile != "" {
		log.Info().Str("path", config.ACMIFile).Msg("opening ACMI file")
		tacviewClient, err = tacview.NewFileClient(
			config.ACMIFile,
			config.Coalition,
			updates,
			fades,
			config.RadarSweepInterval,
		)
	} else {
		log.Info().
			Str("address", config.TelemetryAddress).
			Stringer("timeout", config.TelemetryConnectionTimeout).
			Msg("constructing telemetry client")
		tacviewClient, err = tacview.NewTelemetryClient(
			config.TelemetryAddress,
			config.Callsign,
			config.TelemetryPassword,
			config.Coalition,
			updates,
			fades,
			config.RadarSweepInterval,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	log.Info().Msg("constructing speech-to-text recognizer")
	recognizer := recognizer.NewWhisperRecognizer(config.WhisperModel, config.Callsign)

	log.Info().Msg("constructing text parser")
	parser := parser.New(config.Callsign)

	log.Info().Msg("constructing radar scope")

	rdr := radar.New(config.Coalition, updates, fades, config.MandatoryThreatRadius)
	log.Info().Msg("constructing GCI controller")
	controller := controller.New(
		rdr, srsClient,
		config.Coalition,
		config.SRSFrequency,
		config.PictureBroadcastInterval,
		config.EnableThreatMonitoring,
		config.ThreatMonitoringInterval,
		config.ThreatMonitoringRequiresSRS,
	)

	log.Info().Msg("constructing text composer")
	composer := composer.New(config.Callsign)

	log.Info().Msg("constructing text-to-speech synthesizer")
	synthesizer, err := speakers.NewPiperSpeaker(config.Voice, config.PlaybackSpeed, config.PlaybackPause)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	log.Info().Msg("constructing application")
	app := &app{
		srsClient:     srsClient,
		tacviewClient: tacviewClient,
		recognizer:    recognizer,
		parser:        parser,
		radar:         rdr,
		controller:    controller,
		composer:      composer,
		speaker:       synthesizer,
	}
	return app, nil
}

// Run implements Application.Run.
func (a *app) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) error {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running telemetry client")
		if err := a.tacviewClient.Run(ctx, wg); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("error running telemetry client")
			}
		}

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("updating mission time and bullseye")
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("stopping mission time and bullseye updates due to context cancellation")
				return
			case <-ticker.C:
				missionTime := a.tacviewClient.Time()
				a.radar.SetMissionTime(missionTime)
				for _, coalition := range []coalitions.Coalition{coalitions.Red, coalitions.Blue} {
					bullseye, err := a.tacviewClient.Bullseye(coalition)
					if err != nil {
						log.Warn().Err(err).Msg("error reading bullseye")
					} else {
						a.radar.SetBullseye(bullseye, coalition)
					}
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running SRS client")
		if err := a.srsClient.Run(ctx, wg); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("error running SRS client")
				cancel()
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.recognize(ctx, rxTextChan)
	}()
	log.Info().Msg("starting speech-to-text parsing routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.parse(ctx, rxTextChan, requestChan)
	}()
	log.Info().Msg("starting radar scope routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.radar.Run(ctx, wg)
	}()
	log.Info().Msg("starting GCI controller routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.control(ctx, wg, requestChan, responseAndCallsChan)
	}()
	log.Info().Msg("starting response composer routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.compose(ctx, responseAndCallsChan, txTextChan)
	}()
	log.Info().Msg("starting speech synthesis routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.synthesize(ctx, txTextChan, txAudioChan)
	}()
	log.Info().Msg("starting radio transmission routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.transmit(ctx, txAudioChan)
	}()

	return nil
}

// recognize runs speech recognition on audio received from SRS and forwards recognized text to the given channel.
func (a *app) recognize(ctx context.Context, out chan<- string) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping speech recognition due to context cancellation")
			return
		case sample := <-a.srsClient.Receive():
			a.recognizeSample(ctx, sample, out)
		}
	}
}

func (a *app) recognizeSample(ctx context.Context, sample audio.Audio, out chan<- string) {
	recogCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	log.Info().Msg("recognizing audio sample")
	start := time.Now()
	text, err := a.recognizer.Recognize(recogCtx, sample)
	if err != nil {
		log.Error().Err(err).Msg("error recognizing audio sample")
	} else if text == "" || text == "[BLANK AUDIO]\n" {
		log.Info().Str("text", text).Msg("unable to recognize any words in audio sample")
	} else {
		log.Info().Stringer("clockTime", time.Since(start)).Str("text", text).Msg("recognized audio")
		out <- text
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
			request := a.parser.Parse(text)
			if request != nil {
				logger.Info().Any("request", request).Msg("parsed text")
				out <- request
			} else {
				logger.Info().Msg("unable to parse text, could be silence, chatter, missing GCI callsign")
			}
		}
	}
}

// control routes requests to GCI controller handlers.
func (a *app) control(ctx context.Context, wg *sync.WaitGroup, in <-chan any, out chan<- any) {
	log.Info().Msg("running controller")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.controller.Run(ctx, out)
	}()
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
			case *brevity.TripwireRequest:
				logger.Debug().Msg("routing TRIPWIRE request to controller")
				a.controller.HandleTripwire(request)
			case *brevity.UnableToUnderstandRequest:
				logger.Debug().Msg("routing unable to understand request to controller")
				a.controller.HandleUnableToUnderstand(request)
			default:
				logger.Error().Any("request", brev).Msg("unable to route request to handler")
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
			logger := log.With().Type("type", call).Any("params", call).Logger()
			logger.Info().Msg("composing brevity call")
			var response composer.NaturalLanguageResponse
			switch c := call.(type) {
			case brevity.AlphaCheckResponse:
				logger.Debug().Msg("composing ALPHA CHECK call")
				response = a.composer.ComposeAlphaCheckResponse(c)
			case brevity.BogeyDopeResponse:
				logger.Debug().Msg("composing BOGEY DOPE call")
				response = a.composer.ComposeBogeyDopeResponse(c)
			case brevity.DeclareResponse:
				logger.Debug().Msg("composing DECLARE call")
				response = a.composer.ComposeDeclareResponse(c)
			case brevity.FadedCall:
				logger.Debug().Msg("composing FADED call")
				response = a.composer.ComposeFadedCall(c)
			case brevity.NegativeRadarContactResponse:
				logger.Debug().Msg("composing NEGATIVE RADAR CONTACT call")
				response = a.composer.ComposeNegativeRadarContactResponse(c)
			case brevity.PictureResponse:
				logger.Debug().Msg("composing PICTURE call")
				response = a.composer.ComposePictureResponse(c)
			case brevity.RadioCheckResponse:
				logger.Debug().Msg("composing RADIO CHECK call")
				response = a.composer.ComposeRadioCheckResponse(c)
			case brevity.SnaplockResponse:
				logger.Debug().Msg("composing SNAPLOCK call")
				response = a.composer.ComposeSnaplockResponse(c)
			case brevity.SpikedResponse:
				logger.Debug().Msg("composing SPIKED call")
				response = a.composer.ComposeSpikedResponse(c)
			case brevity.TripwireResponse:
				logger.Debug().Msg("composing TRIPWIRE call")
				response = a.composer.ComposeTripwireResponse(c)
			case brevity.SunriseCall:
				logger.Debug().Msg("composing SUNRISE call")
				response = a.composer.ComposeSunriseCall(c)
			case brevity.ThreatCall:
				logger.Debug().Msg("composing THREAT call")
				response = a.composer.ComposeThreatCall(c)
			case brevity.SayAgainResponse:
				logger.Debug().Msg("composing SAY AGAIN call")
				response = a.composer.ComposeSayAgainResponse(c)
			default:
				logger.Debug().Msg("unable to route call to composition")
			}

			if response.Speech == "" && response.Subtitle == "" {
				logger.Warn().Msg("natural language response is empty")
			} else {
				logger.Info().Str("speech", response.Speech).Str("subtitle", response.Subtitle).Msg("composed brevity call")
				out <- response
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
			start := time.Now()
			audio, err := a.speaker.Say(response.Speech)
			if err != nil {
				log.Error().Err(err).Msg("error synthesizing speech")
			} else {
				if len(audio) == 0 {
					log.Warn().Msg("synthesized audio is empty")
				} else {
					log.Info().Stringer("clockTime", time.Since(start)).Msg("synthesized audio")
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
				log.Info().Msg("transmitting audio")
			}
			a.srsClient.Transmit(audio)
		}
	}
}
