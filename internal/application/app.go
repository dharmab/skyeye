// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	tacview "github.com/dharmab/skyeye/pkg/tacview/client"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
)

// Application is the interface for running the SkyEye application.
type Application interface {
	// Run runs the SkyEye application. It should be called exactly once.
	Run(context.Context, context.CancelFunc, *sync.WaitGroup) error
}

// app implements the Application.
type app struct {
	// callsign of the GCI controller
	callsign string
	// srsClient is a SimpleRadio Standalone client
	srsClient simpleradio.Client
	// tacviewClient streams ACMI data
	tacviewClient tacview.Client
	// recognizer provides speech-to-text recognition
	recognizer recognizer.Recognizer
	// parser converts English brevity text to internal representations
	parser parser.Parser
	// radar tracks contacts and provides geometric computations
	radar radar.Radar
	// controller publishes responses and calls
	controller controller.Controller
	// composer converts responses and calls from internal representations to English brevity text
	composer composer.Composer
	// speaker provides text-to-speech synthesis
	speaker speakers.Speaker
	// enableTranscriptionLogging controls whether transcriptions are included in logs
	enableTranscriptionLogging bool
	// tracers are destinations where traces are sent when tracing is enabled
	tracers []traces.Tracer
}

// NewApplication constructs a new Application.
func NewApplication(ctx context.Context, config conf.Configuration) (Application, error) {
	starts := make(chan sim.Started)
	updates := make(chan sim.Updated)
	fades := make(chan sim.Faded)

	radios := make([]srs.Radio, 0, len(config.SRSFrequencies))
	for _, radioFrequency := range config.SRSFrequencies {
		radios = append(radios, srs.Radio{
			Frequency:        radioFrequency.Frequency.Hertz(),
			Modulation:       radioFrequency.Modulation,
			ShouldRetransmit: true,
		})
	}

	log.Info().
		Str("address", config.SRSAddress).
		Stringer("timeout", config.SRSConnectionTimeout).
		Str("clientName", config.SRSClientName).
		Int("coalitionID", int(config.Coalition)).
		Int("modulationID", int(srs.ModulationAM)).
		Msg("constructing SRS client")
	srsClient, err := simpleradio.NewClient(srs.ClientConfiguration{
		Address:                   config.SRSAddress,
		ConnectionTimeout:         config.SRSConnectionTimeout,
		ClientName:                config.SRSClientName,
		ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
		Coalition:                 config.Coalition,
		Radios:                    radios,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	var tacviewClient tacview.Client
	if config.ACMIFile != "" {
		log.Info().Str("path", config.ACMIFile).Msg("opening ACMI file")
		tacviewClient, err = tacview.NewFileClient(
			config.ACMIFile,
			config.Coalition,
			starts,
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
			starts,
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
	parser := parser.New(config.Callsign, config.EnableTranscriptionLogging)

	log.Info().Msg("constructing radar scope")

	rdr := radar.New(config.Coalition, starts, updates, fades, config.MandatoryThreatRadius)
	log.Info().Msg("constructing GCI controller")
	controller := controller.New(
		rdr,
		srsClient,
		config.Coalition,
		config.EnableAutomaticPicture,
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

	tracers := make([]traces.Tracer, 0)
	if config.EnableTracing {
		log.Info().Msg("constructing tracers")
		tracers = append(tracers, traces.NewLogTracer())
		if config.DiscordWebhookID != "" && config.DiscorbWebhookToken != "" {
			discordWebhook, err := traces.NewDiscordWebhook(config.DiscordWebhookID, config.DiscorbWebhookToken)
			if err != nil {
				return nil, fmt.Errorf("failed to construct application: %w", err)
			}
			tracers = append(tracers, discordWebhook)
		}
	}

	log.Info().Msg("constructing application")
	app := &app{
		callsign:                   config.Callsign,
		enableTranscriptionLogging: config.EnableTranscriptionLogging,
		srsClient:                  srsClient,
		tacviewClient:              tacviewClient,
		recognizer:                 recognizer,
		parser:                     parser,
		radar:                      rdr,
		controller:                 controller,
		composer:                   composer,
		speaker:                    synthesizer,
		tracers:                    tracers,
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
				a.updateMissionTime()
				a.updateBullseyes()
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

	rxTextChan := make(chan Message[string])
	requestChan := make(chan Message[any])
	callChan := make(chan controller.Call)
	txTextChan := make(chan Message[composer.NaturalLanguageResponse])
	txAudioChan := make(chan Message[simpleradio.Audio])

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
		a.control(ctx, wg, requestChan, callChan)
	}()
	log.Info().Msg("starting response composer routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.compose(ctx, callChan, txTextChan)
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

// updateMissionTime updates the mission time on the radar.
func (a *app) updateMissionTime() {
	missionTime := a.tacviewClient.Time()
	a.radar.SetMissionTime(missionTime)
}

// updateBullseyes updates the positions of the bullseyes on the radar.
func (a *app) updateBullseyes() {
	for _, coalition := range []coalitions.Coalition{coalitions.Red, coalitions.Blue} {
		bullseye, err := a.tacviewClient.Bullseye(coalition)
		if err != nil {
			log.Warn().Err(err).Msg("error reading bullseye")
		} else {
			a.radar.SetBullseye(bullseye, coalition)
		}
	}
}

// trace the given request context using all configured tracers.
func (a *app) trace(ctx context.Context) {
	ctx = traces.WithCallsign(ctx, a.callsign)
	if !a.enableTranscriptionLogging {
		ctx = traces.WithoutRequestText(ctx)
	}
	for _, tracer := range a.tracers {
		tracer.Trace(ctx)
	}
}
